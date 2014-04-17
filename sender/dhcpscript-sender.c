/*
 * Copyright 2014 Jeffrey Clark. All rights reserved.
 * License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses.gpl.html>.
 * This is free software: you are free to change and redistribute it.
 * There is NO WARRANTY, to the extent permitted by law.
 *
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <syslog.h>

#define MAXLINE 2048
#define REMOTE_IP "10.1.1.11"
#define REMOTE_PORT "9999"

extern char **environ;

void mylog(int priority, const char *format, ...) {
    va_list args;
    va_start(args, format);
    if (isatty(fileno(stdin))) {
        vprintf(format, args);
    } else {
        vsyslog(priority, format, args);
    }
    va_end(args);
}

int main(int argc, char** argv)
{
    int sockfd, n;
    char buff[MAXLINE], recvline[MAXLINE];
    struct sockaddr_in servaddr;
    char request[MAXLINE] = "/dnsmasq";

    if (argc < 4 || argc > 5)
    {
        mylog(LOG_ERR, "usage: %s <action> <mac> <ip> [host]\n", argv[0]);
        exit(EXIT_FAILURE);
    }

    if ( (sockfd = socket(AF_INET,SOCK_STREAM,IPPROTO_TCP)) == -1 )
    {
        mylog(LOG_ERR, "socket: error");
        exit(EXIT_FAILURE);
    }

    sprintf(&request[strlen(request)], "/%s/%s?id=%s", argv[1], argv[3], argv[2]);
    if (argc == 5)
    {
        sprintf(&request[strlen(request)], "&host=%s", argv[4]);
    }

    char **env;
    char *token;
    char delims[] = "=";

    for (env = environ; *env != 0; env++)
    {
        char* curenv = *env;
        token = strtok( curenv, delims );
        if (strncmp(token, "DNSMASQ_", 7) == 0)
        {
            for (n = 0; token[n]; n++)
            {
                if (token[n]<91 && token[n] > 64)
                {
                    token[n] = token[n] + 32;
                }
            }

            sprintf(&request[strlen(request)], "&%s=", token + 8);
            token = strtok( NULL, delims );
            if (token != NULL)
            {
                for (n = 0; token[n]; n++)
                {
                    if (token[n] == ' ') token[n] = '+';
                }
                sprintf(&request[strlen(request)], "%s", token);
            }
        }
    }

    bzero(&servaddr,sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(atoi(REMOTE_PORT));

    if(inet_pton(AF_INET,REMOTE_IP,&servaddr.sin_addr) < 0)
    {
        mylog(LOG_ERR, "assigned port invalid (%s)", REMOTE_PORT);
        exit(EXIT_FAILURE);
    }
    if( connect(sockfd,(struct sockaddr *)&servaddr,sizeof(servaddr)) == -1)
    {
        mylog(LOG_ERR, "failed to connect to %s:%s", REMOTE_IP, REMOTE_PORT);
        exit(EXIT_FAILURE);
    }

    bzero(&buff,sizeof(buff));

    sprintf(buff, "GET %s HTTP/1.0\r\nAccept: */*\r\n\r\n", request);

    if(write(sockfd,buff,strlen(buff)+1) == -1)
    {
        mylog(LOG_ERR, "write error sending request: %s", request);
        exit(EXIT_FAILURE);
    }

    while ((n =read(sockfd,recvline,sizeof(recvline))) > 0)
    {
        recvline[n] = 0;
        if(fputs(recvline,stdout) == EOF)
        {
            mylog(LOG_ERR, "error reading response to: %s", request);
            exit(EXIT_FAILURE);
        }
    }

    mylog(LOG_DEBUG, "%s", recvline);

    return (EXIT_SUCCESS);
}
