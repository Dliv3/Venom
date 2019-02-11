#!/usr/bin/env python
# -*- coding: utf-8 -*-

# a simple script to start/stop iptables port reuse.

# example:
#   python port_reuse.py --start --rhost 192.168.1.2 --rport 80
#   python port_reuse.py --stop --rhost 192.168.1.2 --rport 80

import socket
import argparse
import sys

parser = argparse.ArgumentParser(description='start/stop iptables port reuse')
parser.add_argument('--start', help='start port reusing', action='store_true')
parser.add_argument('--stop', help='stop port reusing', action='store_true')
parser.add_argument('--rhost', help='remote host', dest='ip')
parser.add_argument('--rport', help='remote port', dest='port')

START_PORT_REUSE = "venomcoming"
STOP_PORT_REUSE = "venomleaving"

options = parser.parse_args()    

data = ""

if options.start:
    data = START_PORT_REUSE
elif options.stop:
    data = STOP_PORT_REUSE
else:
    parser.print_help()
    sys.exit(0)

try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(2)
    s.connect((options.ip, int(options.port)))
    s.send(data)
except:
    print "[-]Connect to target host error."

try:
    s.recv(1024)
except:
    pass

s.close()

print "[+]Done!"
