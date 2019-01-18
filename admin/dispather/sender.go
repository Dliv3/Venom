package dispather

import (
	"fmt"
	"sync"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
	"gitlab.com/Dliv3/Venom/util"
)

// SendSyncCmd 发送同步网络拓扑的命令
func SendSyncCmd() {

	// 重新初始化网络拓扑，这样当有节点断开时网络拓扑会实时改变
	node.GNetworkTopology.InitNetworkMap()

	for i := range node.Nodes {
		// 向直连的节点发送SYNC数据包，同步网络拓扑
		// 目标节点会递归处理SYNC数据包，以获得全网拓扑
		if node.Nodes[i].DirectConnection {

			// 构造命令数据包
			packetHeader := protocol.PacketHeader{
				Separator: global.PROTOCOL_SEPARATOR,
				SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
				DstHashID: utils.UUIDToArray32(node.Nodes[i].HashID),
				CmdType:   protocol.SYNC,
			}
			// 生成本节点网络拓扑数据
			networkMap := node.GNetworkTopology.GenerateNetworkMapData()
			syncPacket := protocol.SyncPacket{
				NetworkMapLen: uint64(len(networkMap)),
				NetworkMap:    networkMap,
			}

			// 发送命令数据包
			node.Nodes[i].WritePacket(packetHeader, syncPacket)

			// 读取返回数据包
			node.CurrentNode.CommandBuffers[protocol.SYNC].ReadPacket(&packetHeader, &syncPacket)

			// 解析网络拓扑数据包
			node.GNetworkTopology.ResolveNetworkMapData(syncPacket.NetworkMap)
		}
	}

	// 生成路由表
	node.GNetworkTopology.UpdateRouteTable()

	// 生成节点信息
	node.GNodeInfo.UpdateNoteInfo()

	// fmt.Println(node.CurrentNode.HashID)
	// fmt.Println(node.GNetworkTopology.RouteTable)
	// fmt.Println(node.GNetworkTopology.NetworkMap)

	// 创建Node结构体
	for key, value := range node.GNetworkTopology.RouteTable {
		if _, ok := node.Nodes[key]; !ok {
			node.Nodes[key] = &node.Node{
				HashID:               key,
				Conn:                 node.Nodes[value].Conn,
				ConnReadLock:         &sync.Mutex{},
				ConnWriteLock:        &sync.Mutex{},
				Socks5SessionIDLock:  &sync.Mutex{},
				Socks5DataBufferLock: &sync.RWMutex{},
			}
		}
	}

}

// SendListenCmd 发送监听端口命令
func SendListenCmd(peerNode *node.Node, port uint16) {
	listenPacketCmd := protocol.ListenPacketCmd{
		Port: port,
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: util.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: util.UUIDToArray32(global.CurrentPeerNodeHashID),
		CmdType:   protocol.LISTEN,
	}

	peerNode.WritePacket(packetHeader, listenPacketCmd)

	var listenPacketRet protocol.ListenPacketRet
	node.CurrentNode.CommandBuffers[protocol.LISTEN].ReadPacket(&packetHeader, &listenPacketRet)

	if listenPacketRet.Success == 1 {
		fmt.Println("listen local port success!")
	} else {
		fmt.Println("listen local port failed!")
		fmt.Println(string(listenPacketRet.Msg))
	}
}

func SendConnectCmd(peerNode *node.Node, ip string, port uint16) {
	// connectPacketCmd := protocol.ConnectPacketCmd{
	// 	IP:   util.IpToUint32(net.ParseIP(ip)),
	// 	Port: port,
	// }
	// dataLen, _ := util.PacketSize(connectPacketCmd)
	// packetHeader := protocol.PacketHeader{
	// 	Separator: protocol.SEPARATOR,
	// 	SrcHashID: util.UUIDToArray32(node.CurrentNode.HashID),
	// 	DstHashID: util.UUIDToArray32(node.CurrentPeerNodeHashID),
	// 	CmdType:   protocol.CONNECT,
	// 	DataLen:   dataLen,
	// }

	// peerNode.WritePacket(packetHeader, connectPacketCmd)

	// packet, _ := node.CommandBuffers[protocol.CONNECT].ReadLowLevelPacket()

	// var connectPacketRet protocol.ConnectPacketRet
	// packet.ResolveData(&connectPacketRet)

	// if connectPacketRet.Success == 1 {
	// 	fmt.Println("connect to remote port success!")
	// } else {
	// 	fmt.Println("connect to remote port failed!")
	// 	fmt.Println(string(connectPacketRet.Msg))
	// }
}
