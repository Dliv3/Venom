package cli

import (
	"flag"
	"fmt"
	"os"
)

// COMMAND LINE INTERFACE

const (
	LISTEN_MODE  = 1
	CONNECT_MODE = 2
)

type Option struct {
	LocalPort  int
	RemoteIP   string
	RemotePort int
	Mode       int
}

// Args
var Args Option

func init() {
	flag.IntVar(&Args.LocalPort, "l", 0, "Listen a local `port`.")
	flag.StringVar(&Args.RemoteIP, "c", "", "Remote `ip` address.")
	flag.IntVar(&Args.RemotePort, "p", 0, "The `port` on remote host.")

	// change default Usage
	flag.Usage = usage
}

func usage() {
	ShowBanner()
	fmt.Fprintf(os.Stderr, `Venom version: 1.0

Usage:
	$ ./venom_admin -l [lport]
	$ ./venom_admin -c [rhost] -p [rport]

Options:
`)
	flag.PrintDefaults()
}

// ParseArgs is a function aim to parse the command line args
func ParseArgs() {
	flag.Parse()

	if Args.LocalPort == 0 && Args.RemoteIP != "" && Args.RemotePort != 0 {
		// connect to remote port
		Args.Mode = CONNECT_MODE
	} else if Args.LocalPort != 0 && Args.RemoteIP == "" && Args.RemotePort == 0 {
		// listen a local port
		Args.Mode = LISTEN_MODE
	} else {
		// error
		flag.Usage()
		os.Exit(0)
	}
}

func PrintBanner(data string) {
	fmt.Printf("\x1b[0;34m%s \x1b[0m", data)
	fmt.Println()
}

func ShowBanner() {
	PrintBanner(` _    _                             
| |  | |   author: Dlive   v1.0
| |  | |  ____  ____    ___   ____  
 \ \/ /  / _  )|  _ \  / _ \ |    \ 
  \  /  ( (/ / | | | || |_| || | | |
   \/    \____)|_| |_| \___/ |_|_|_| 
								`)
}

// ShowUsage
func ShowUsage() {
	fmt.Println(`
  help                                     Help information.
  exit                                     Exit.
  show                                     Display network topology.
  setdes     [id] [info]                   Add a description to the target node.
  getdes     [id]                          View description of the target node.
  goto       [id]                          Select id as the target node.
  listen     [port]                        Listen on a port on the target node.
  connect    [ip] [port]                   Connect to a new node through current node.
  sshconnect [user@ip:port] [dport]        Connect to a new node through ssh tunnel.
  shell                                    Start an interactive shell on the target node.
  upload     [local_file]  [remote_file]   Upload file to the target node.
  download   [remote_file]  [local_file]   Download file from the target node.
  socks      [lport]                       Start a socks server.
  lforward   [lhost] [sport] [dport]       Forward a local sport to a remote dport.
  rforward   [rhost] [sport] [dport]       Forward a remote sport to a local dport.
`)
}
