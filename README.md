_General_

Source to the programs I use on my home network for relaying dhcp leases from dnsmasq (sender) to my Samba 4 AD DC runnnig bind 9.

_Sender_

Written in C because I needed something efficient enough to run on SOHO linux routers (Tomato, DD-WRT, OpenWRT, etc).  Still need to make the end point configurable, currently it's hard-coded in the source to 10.1.1.11:9999.  I will probably change it to look at an environment variable set by dnsmasq in the future.

_Receiver_

Seriously [Go](http://golang.org/)?? ... why not.  Rough around the edges, but I have been running it for a few days without issue at home.  Simply does a kerberized nsupdate request, something you generally cannot do on SOHO routers.

_License_

Copyright 2014 Jeffrey Clark. All rights reserved.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses.gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
