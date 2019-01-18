package main

import (
	"fmt"
	"runtime"

	"github.com/Dliv3/Venom/agent/cli"
	"github.com/Dliv3/Venom/agent/dispather"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/node"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cli.ParseArgs()

	node.CurrentNode.IsAdmin = 0
	node.CurrentNode.InitCommandBuffer()

	fmt.Println(node.CurrentNode.HashID)

	dispather.InitAgentHandler()

	if cli.Args.Mode == cli.LISTEN_MODE {
		// 监听端口
		if cli.Args.PortReuse {
			netio.Init(
				"listen",
				fmt.Sprintf("%s:%d", cli.Args.LocalIP, uint16(cli.Args.LocalPort)),
				dispather.AgentServer, true)
		} else {
			netio.Init(
				"listen",
				fmt.Sprintf("0.0.0.0:%d", uint16(cli.Args.LocalPort)),
				dispather.AgentServer, false)
		}
	} else {
		// 连接端口
		netio.Init(
			"connect",
			fmt.Sprintf("%s:%d", cli.Args.RemoteIP, uint16(cli.Args.RemotePort)),
			dispather.AgentClient, false)
	}

	var exit string
	for exit != "exit" {
		fmt.Scanln(&exit)
	}
}
