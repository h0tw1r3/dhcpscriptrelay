/*
 * Copyright 2014 Jeffrey Clark. All rights reserved.
 * License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses.gpl.html>.
 * This is free software: you are free to change and redistribute it.
 * There is NO WARRANTY, to the extent permitted by law.
 *
 */

package main

import (
	"bytes"
	"log"
	"fmt"
	"os"
	"os/exec"
	"net"
	"strings"
	"text/template"
	"github.com/hoisie/web"
	"code.google.com/p/gcfg"
)

type Config struct {
	Kerberos struct {
		Realm string
		Principal string
		Keytab string
		Krb5cc string
		Klist string
		Klistarg string
		Kinit string
	}
	Dns struct {
		Defaultdomain string
		Nsrvs string
		Ttl int
		Notxtrrs bool
		Proto string
		Nsupdateflags string
	}
	Service struct {
		Ip string
		Port string
		Debug bool
	}
}

type Nsupdate struct {
	Server string
	Ip string
	Host string
	Domain string
	Rr string
	Ptr string
	Ttl int
	Mac string
}

var cfg Config
var tmpl *template.Template

func reverse(s string) string {
	n := len(s)
	runes := make([]rune, n)
	for _, rune := range s {
		n--
		runes[n] = rune
	}
	return string(runes[n:])
}

func init() {
	err := gcfg.ReadFileInto(&cfg, "dhcplistener.ini")
	if err != nil {
		log.Fatalf("failed to parse ini: %s", err)
	}

	os.Setenv("KRB5_KTNAME", cfg.Kerberos.Keytab)
	os.Setenv("KRB5CCNAME", cfg.Kerberos.Krb5cc)

	tmpl = template.Must(template.ParseFiles("dhcplistener.nsupdate.add", "dhcplistener.nsupdate.del"))
}

func dnsmasq(ctx *web.Context, action string, ip string) string {
	actions := map[string]bool {
		"add": true,
		"del": true,
		"old": true,
		"init": false,
		"tftp": false,
	}
	response := ""

	if actions[action] {
		if !kerberos() {
			log.Printf("Unable to process %s", action)
			ctx.NotFound("That sucks")
			return ""
		}

		tv := Nsupdate{
			Server: cfg.Dns.Nsrvs,
			Ip: ip,
			Host: ctx.Params["host"],
			Domain: ctx.Params["domain"],
			Ttl: cfg.Dns.Ttl,
			Mac: ctx.Params["mac"],
		}

		IP := net.ParseIP(ip)
		if IP.To4() == nil && IP.To16() != nil {
			tv.Rr = "AAA"
			var expanded string
			for u := 0; u < len(IP); u++ {
				x := fmt.Sprintf("%x", IP[u])
				if len(x) != 2 {
					expanded += "0"
				}
				expanded += x
				if u%2 != 0 && u < (len(IP)-1) {
					expanded += ":"
				}
			}
			tv.Ip = fmt.Sprintf("%s", expanded)
			for _, r := range reverse(expanded) {
				c := string(r)
				if c != ":" {
					tv.Ptr += fmt.Sprintf("%s.", c)
				}
			}
			tv.Ptr += "ipv6.arpa"
		} else {
			s := strings.Split(ip, ".")
			tv.Rr = "A"
			tv.Ptr = fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa", s[3], s[2], s[1], s[0])
		}

		if len(tv.Host) < 2 {
			log.Print("nil host!")
			addr, _ := net.LookupAddr(ip)
			if addr != nil && cfg.Service.Debug {
				log.Printf("LookupAddr: %d%s", len(addr), addr)
			}
		}

		var output bytes.Buffer
		var stderr bytes.Buffer

		nsupdate := exec.Command("nsupdate", cfg.Dns.Nsupdateflags)
		nsin, err := nsupdate.StdinPipe()
		if err != nil {
			log.Fatalf("unable to create pipe: %s", err)
		}
		nsupdate.Stdout, nsupdate.Stderr = &output, &stderr

		nsupdate.Start()

		if action == "add" || action == "old" {
			err = tmpl.Execute(nsin, tv)
			if err != nil {
				log.Fatalf("template execution failed: %s", err)
			}

			response = fmt.Sprintf("%s %s\n", ip, action)
		}

		if action == "del" {
			err := tmpl.ExecuteTemplate(os.Stderr, "dhcplistener.nsupdate.del", tv)
			if err != nil {
				log.Fatalf("template execution failed: %s", err)
			}

			response = fmt.Sprintf("%s %s\n", ip, action)
		}

		nsin.Close()
		nsupdate.Wait()

		if cfg.Service.Debug {
			log.Printf("%s", output)
			log.Printf("%s", stderr)
		}

		return response
	}

	ctx.NotFound("That sucks")
	return ""
}

func kerberos() bool {
	klist := exec.Command(cfg.Kerberos.Klist, cfg.Kerberos.Klistarg)
	klistargs := strings.Split(cfg.Kerberos.Klistarg, " ")
	if len(klistargs) > 1 {
		klist.Args = klistargs
	}
	klisterr := klist.Run()

	if klisterr != nil {
		log.Printf("klist (%s) failed: %s", cfg.Kerberos.Klistarg, klisterr)
		kinit := exec.Command(cfg.Kerberos.Kinit, "-F", "-k", cfg.Kerberos.Principal)

		var output bytes.Buffer
		var stderr bytes.Buffer

		kinit.Stdout, kinit.Stderr = &output, &stderr
		kiniterr := kinit.Run()

		if kiniterr != nil {
			log.Printf("kinit failed: %s", kiniterr)
			log.Printf("output: %s", output)
			log.Printf("stderr: %s", stderr)
			return false
		}
	}
	return true
}

func main() {
	web.Get("/dnsmasq/([^/]*)/([^/]*)", dnsmasq)
	web.Run(fmt.Sprintf("%s:%s", cfg.Service.Ip, cfg.Service.Port))
}
