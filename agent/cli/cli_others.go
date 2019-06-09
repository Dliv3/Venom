// +build windows darwin
// +build amd64 386

package cli

import (
	"flag"
	"fmt"
	"os"
)

// COMMAND LINE INTERFACE

func init() {
	// 不加-h选项也可以正确显示帮助信息
	// flag.BoolVar(&help, "h", false, "help")

	flag.IntVar(&Args.LocalPort, "lport", 0, "Listen a local `port`.")
	flag.StringVar(&Args.LocalIP, "lhost", "", "Local `ip` address.")
	flag.IntVar(&Args.RemotePort, "rport", 0, "The `port` on remote host.")
	flag.StringVar(&Args.RemoteIP, "rhost", "", "Remote `ip` address.")
	flag.IntVar(&Args.ReusedPort, "reuse-port", 0, "Reused port.")
	flag.StringVar(&Args.Password, "passwd", "", "Password used in encrypted communication.")

	// 改变默认的 Usage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `Venom version: 1.0

Usage:
	$ ./venom_agent -lport [port]
	$ ./venom_agent -rhost [ip] -rport      [port]
	$ ./venom_agent -lhost [ip] -reuse-port [port]

Options:
`)
	flag.PrintDefaults()
}

// ParseArgs is a function aim to parse the command line args
func ParseArgs() {
	flag.Parse()

	if Args.RemoteIP != "" && Args.RemotePort != 0 && Args.LocalPort == 0 &&
		Args.ReusedPort == 0 && Args.LocalIP == "" {
		// connect to remote port
		Args.Mode = CONNECT_MODE
	} else if Args.LocalPort != 0 && Args.RemoteIP == "" && Args.RemotePort == 0 &&
		Args.ReusedPort == 0 && Args.LocalIP == "" {
		// listen a local port
		Args.Mode = LISTEN_MODE
	} else if Args.ReusedPort != 0 && Args.LocalIP != "" && Args.LocalPort == 0 &&
		Args.RemoteIP == "" && Args.RemotePort == 0 {
		Args.Mode = LISTEN_MODE
		Args.PortReuseMethod = SOCKET_REUSE_METHOD
	} else {
		// error
		flag.Usage()
		os.Exit(0)
	}
}
