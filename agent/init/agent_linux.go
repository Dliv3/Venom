// +build linux
// +build amd64 386

package init

import (
	"fmt"
	"log"

	"github.com/Dliv3/Venom/agent/cli"
	"github.com/Dliv3/Venom/agent/dispather"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/utils"
)

func InitNode() {

	if cli.Args.Mode == cli.LISTEN_MODE {
		// 端口复用
		if cli.Args.ReusedPort != 0 {
			switch cli.Args.PortReuseMethod {
			case cli.SOCKET_REUSE_METHOD:
				netio.InitNode(
					"listen",
					fmt.Sprintf("%s:%d", cli.Args.LocalIP, uint16(cli.Args.ReusedPort)),
					dispather.AgentServer, true, uint16(cli.Args.ReusedPort))
			case cli.IPTABLES_METHOD:
				defer utils.DeletePortReuseRules(uint16(cli.Args.LocalPort), uint16(cli.Args.ReusedPort))
				err := utils.SetPortReuseRules(uint16(cli.Args.LocalPort), uint16(cli.Args.ReusedPort))
				if err != nil {
					log.Println("[-]Add iptables rules error:", err)
					return
				}
				netio.InitNode(
					"listen",
					fmt.Sprintf("0.0.0.0:%d", uint16(cli.Args.LocalPort)),
					dispather.AgentServer, true, uint16(cli.Args.ReusedPort))
			}
		} else {
			netio.InitNode(
				"listen",
				fmt.Sprintf("0.0.0.0:%d", uint16(cli.Args.LocalPort)),
				dispather.AgentServer, false, 0)
		}
	} else {
		// 连接端口
		netio.InitNode(
			"connect",
			fmt.Sprintf("%s:%d", cli.Args.RemoteIP, uint16(cli.Args.RemotePort)),
			dispather.AgentClient, false, 0)
	}

	select {}
}
