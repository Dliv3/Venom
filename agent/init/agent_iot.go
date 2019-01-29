// +build !386
// +build !amd64

package init

import (
	"fmt"

	"github.com/Dliv3/Venom/agent/cli"
	"github.com/Dliv3/Venom/agent/dispather"
	"github.com/Dliv3/Venom/netio"
)

func InitNode() {
	if cli.Args.Mode == cli.LISTEN_MODE {
		// 监听端口
		netio.InitNode(
			"listen",
			fmt.Sprintf("0.0.0.0:%d", uint16(cli.Args.LocalPort)),
			dispather.AgentServer, false, 0)
	} else {
		// 连接端口
		netio.InitNode(
			"connect",
			fmt.Sprintf("%s:%d", cli.Args.RemoteIP, uint16(cli.Args.RemotePort)),
			dispather.AgentClient, false, 0)
	}

	var exit string
	for exit != "exit" {
		fmt.Scanln(&exit)
	}
}
