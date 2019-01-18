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
	LocalIP    string
	PortReuse  bool
	// 0 默认值，表示参数解析错误，无法设置模式
	// mode 1 listen a local port
	// mode 2 connect to remote port
	Mode int
}

// Args
var Args Option

func init() {
	// 不加-h选项也可以正确显示帮助信息
	// flag.BoolVar(&help, "h", false, "help")

	flag.IntVar(&Args.LocalPort, "l", 0, "Listen a local `port`.")
	flag.StringVar(&Args.LocalIP, "h", "", "Local `ip` address.")
	flag.BoolVar(&Args.PortReuse, "reuse-port", false, "Reuse port.")
	flag.StringVar(&Args.RemoteIP, "c", "", "Remote `ip` address.")
	flag.IntVar(&Args.RemotePort, "p", 0, "The `port` on remote host.")

	// 改变默认的 Usage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `Venom version: 1.0

Usage:
	$ ./venom_agent -l [lport]
	$ ./venom_agent -c [rhost] -p [rport]
	$ ./venom_agent -h [lhost] -l [lport] -reuse-port

Options:
`)
	flag.PrintDefaults()
}

// ParseArgs is a function aim to parse the command line args
func ParseArgs() {
	flag.Parse()

	if Args.LocalPort == 0 && Args.RemoteIP != "" && Args.RemotePort != 0 &&
		!Args.PortReuse && Args.LocalIP == "" {
		// connect to remote port
		Args.Mode = CONNECT_MODE
	} else if Args.LocalPort != 0 && Args.RemoteIP == "" && Args.RemotePort == 0 &&
		!Args.PortReuse && Args.LocalIP == "" {
		// listen a local port
		Args.Mode = LISTEN_MODE
	} else if Args.LocalPort != 0 && Args.LocalIP != "" && Args.PortReuse &&
		Args.RemoteIP == "" && Args.RemotePort == 0 {
		Args.Mode = LISTEN_MODE
	} else {
		// error
		flag.Usage()
		os.Exit(0)
	}
}
