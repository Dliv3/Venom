package dispather

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
	"gitlab.com/Dliv3/Venom/util"
)

var ERR_UNKNOWN_CMD = errors.New("unknown command type")
var ERR_PROTOCOL_SEPARATOR = errors.New("unknown separator")
var ERR_TARGET_NODE = errors.New("can not find target node")

// AgentClient Admin节点作为Client
func AgentClient(conn net.Conn) {
	result, peerNode := node.ClentInitConnection(conn)
	if result {
		log.Println("[+]Connect to a new node success")
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

// AgentServer Admin节点作为Server
func AgentServer(conn net.Conn) {
	log.Println("[+]Remote Connection: ", conn.RemoteAddr())
	result, peerNode := node.ServerInitConnection(conn)
	if result {
		log.Println("[+]A new node connect to this node success")
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

// InitAgentHandler Agent处理Admin发出的命令
func InitAgentHandler() {
	go handleSyncCmd()
	go handleListenCmd()
	go handleConnectCmd()
}

func handleSyncCmd() {
	for {
		// fmt.Println("Nodes", node.Nodes)

		var packetHeader protocol.PacketHeader
		var syncPacket protocol.SyncPacket
		node.CurrentNode.CommandBuffers[protocol.SYNC].ReadPacket(&packetHeader, &syncPacket)

		// 重新初始化网络拓扑，这样当有节点断开时网络拓扑会实时改变
		node.GNetworkTopology.InitNetworkMap()

		node.GNetworkTopology.ResolveNetworkMapData(syncPacket.NetworkMap)

		// 通信的对端节点
		var peerNodeID = utils.Array32ToUUID(packetHeader.SrcHashID)

		// nextNode为下一跳
		nextNode := node.Nodes[node.GNetworkTopology.RouteTable[peerNodeID]]

		// 递归向其他节点发送sync同步路由表请求
		for i := range node.Nodes {
			if node.Nodes[i].HashID != peerNodeID && node.Nodes[i].DirectConnection {
				tempPacketHeader := protocol.PacketHeader{
					Separator: global.PROTOCOL_SEPARATOR,
					SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
					DstHashID: utils.UUIDToArray32(node.Nodes[i].HashID),
					CmdType:   protocol.SYNC,
				}
				networkMap := node.GNetworkTopology.GenerateNetworkMapData()
				tempSyncPacket := protocol.SyncPacket{
					NetworkMapLen: uint64(len(networkMap)),
					NetworkMap:    networkMap,
				}

				node.Nodes[i].WritePacket(tempPacketHeader, tempSyncPacket)

				node.CurrentNode.CommandBuffers[protocol.SYNC].ReadPacket(&tempPacketHeader, &tempSyncPacket)

				node.GNetworkTopology.ResolveNetworkMapData(tempSyncPacket.NetworkMap)
			}
		}

		// 生成路由表
		node.GNetworkTopology.UpdateRouteTable()

		// fmt.Println("RouteTable", node.GNetworkTopology.RouteTable)

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

		// // 生成节点信息
		// node.GNodeInfo.UpdateNoteInfo()

		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
			CmdType:   protocol.SYNC,
		}
		networkMap := node.GNetworkTopology.GenerateNetworkMapData()
		syncPacket = protocol.SyncPacket{
			NetworkMapLen: uint64(len(networkMap)),
			NetworkMap:    networkMap,
		}
		nextNode.WritePacket(packetHeader, syncPacket)

		// fmt.Println(node.CurrentNode.HashID)
		// fmt.Println(node.GNetworkTopology.RouteTable)
		// fmt.Println(node.GNetworkTopology.NetworkMap)
	}
}

func handleListenCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var listenPacketCmd protocol.ListenPacketCmd
		node.CurrentNode.CommandBuffers[protocol.LISTEN].ReadPacket(&packetHeader, &listenPacketCmd)

		// adminNode := node.Nodes[node.GNetworkTopology.RouteTable[util.Array32ToUUID(packetHeader.SrcHashID)]]

		// 网络拓扑同步完成之后即可直接使用以及构造好的节点结构体
		adminNode := node.Nodes[util.Array32ToUUID(packetHeader.SrcHashID)]

		err := netio.Init(
			"listen",
			fmt.Sprintf("0.0.0.0:%d", listenPacketCmd.Port),
			AgentServer, false)

		var listenPacketRet protocol.ListenPacketRet
		if err != nil {
			listenPacketRet.Success = 0
			listenPacketRet.Msg = []byte(fmt.Sprintf("%s", err))
		} else {
			listenPacketRet.Success = 1
		}
		listenPacketRet.MsgLen = uint32(len(listenPacketRet.Msg))
		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.LISTEN,
			SrcHashID: util.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeader, listenPacketRet)
	}
}

func handleConnectCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var connectPacketCmd protocol.ConnectPacketCmd

		node.CurrentNode.CommandBuffers[protocol.CONNECT].ReadPacket(&packetHeader, &connectPacketCmd)

		adminNode := node.Nodes[util.Array32ToUUID(packetHeader.SrcHashID)]

		err := netio.Init(
			"connect",
			fmt.Sprintf("%s:%d", util.Uint32ToIp(connectPacketCmd.IP).String(), connectPacketCmd.Port),
			AgentClient, false)

		var connectPacketRet protocol.ConnectPacketRet
		if err != nil {
			connectPacketRet.Success = 0
			connectPacketRet.Msg = []byte(fmt.Sprintf("%s", err))
		} else {
			connectPacketRet.Success = 1
		}
		connectPacketRet.MsgLen = uint32(len(connectPacketRet.Msg))
		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.CONNECT,
			SrcHashID: util.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeader, connectPacketRet)
	}
}
