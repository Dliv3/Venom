package main

import (
	"fmt"
	"runtime"

	"github.com/Dliv3/Venom/admin/cli"
	"github.com/Dliv3/Venom/admin/dispather"
	"github.com/Dliv3/Venom/crypto"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/node"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	cli.ParseArgs()
	fmt.Println("Venom Admin Node Start...")
	cli.ShowBanner()

	// fmt.Println(node.CurrentNode.HashID)

	node.CurrentNode.IsAdmin = 1
	crypto.InitEncryption(cli.Args.Password)
	node.CurrentNode.InitCommandBuffer()
	node.CurrentNode.InitDataBuffer()

	dispather.InitAdminHandler()

	if cli.Args.Mode == cli.CONNECT_MODE {
		netio.InitNode(
			"connect",
			fmt.Sprintf("%s:%d", cli.Args.RemoteIP, uint16(cli.Args.RemotePort)),
			dispather.AdminClient, false, 0)
	} else if cli.Args.Mode == cli.LISTEN_MODE {
		netio.InitNode(
			"listen",
			fmt.Sprintf("0.0.0.0:%d", uint16(cli.Args.LocalPort)),
			dispather.AdminServer, false, 0)
	}
	cli.Interactive()
}
