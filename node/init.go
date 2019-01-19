package node

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

var ERR_UNKNOWN_CMD = errors.New("Unknown command type")

func ServerInitConnection(conn net.Conn) (bool, *Node) {
	// 端口重用模式下发送的一段垃圾数据
	netio.Write(conn, []byte(global.PROTOCOL_FEATURE))

	var PacketHeader protocol.PacketHeader
	netio.ReadPacket(conn, &PacketHeader)

	if PacketHeader.Separator != global.PROTOCOL_SEPARATOR ||
		PacketHeader.CmdType != protocol.INIT {
		log.Println("[-]InitPacket error: separator or cmd type")
		conn.Close()
		return false, nil
	}

	var initPacketCmd protocol.InitPacketCmd
	netio.ReadPacket(conn, &initPacketCmd)

	initPacketRet := protocol.InitPacketRet{
		OsType:  utils.GetSystemType(),
		HashID:  utils.UUIDToArray32(CurrentNode.HashID),
		IsAdmin: 0,
	}
	size, _ := utils.PacketSize(initPacketRet)
	PacketHeader = protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		CmdType:   protocol.INIT,
		DataLen:   size,
	}
	netio.WritePacket(conn, PacketHeader)
	netio.WritePacket(conn, initPacketRet)

	clientNode := &Node{
		HashID:               utils.Array32ToUUID(initPacketCmd.HashID),
		IsAdmin:              initPacketCmd.IsAdmin,
		Conn:                 conn,
		ConnReadLock:         &sync.Mutex{},
		ConnWriteLock:        &sync.Mutex{},
		Socks5SessionIDLock:  &sync.Mutex{},
		Socks5DataBufferLock: &sync.RWMutex{},
		DirectConnection:     true,
	}

	Nodes[utils.Array32ToUUID(initPacketCmd.HashID)] = clientNode
	clientNodeID := utils.Array32ToUUID(initPacketCmd.HashID)
	GNetworkTopology.AddRoute(clientNodeID, clientNodeID)
	GNetworkTopology.AddNetworkMap(CurrentNode.HashID, clientNodeID)
	GNodeInfo.AddNode(clientNodeID)

	return true, clientNode
}

func ClentInitConnection(conn net.Conn) (bool, *Node) {
	// 端口重用模式下发送的一段垃圾数据
	netio.Write(conn, []byte(global.PROTOCOL_FEATURE))

	// Node的初始状态为UNINIT，所以首先CurrentNode会向连接的对端发送init packet
	initPacketCmd := protocol.InitPacketCmd{
		OsType:  utils.GetSystemType(),
		HashID:  utils.UUIDToArray32(CurrentNode.HashID),
		IsAdmin: 0,
	}
	size, _ := utils.PacketSize(initPacketCmd)
	PacketHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		CmdType:   protocol.INIT,
		DataLen:   size,
	}
	netio.WritePacket(conn, PacketHeader)
	netio.WritePacket(conn, initPacketCmd)

	// 读取返回包
	// init阶段可以看做连接建立阶段，双方进行握手后交换信息
	// 所有init阶段无需校验数据包中的DstHashID,因为此时双方还没有获取双方的HashID
	netio.ReadPacket(conn, &PacketHeader)
	if PacketHeader.Separator != global.PROTOCOL_SEPARATOR ||
		PacketHeader.CmdType != protocol.INIT {
		log.Println("[-]InitPacket error: separator or cmd type error")
		conn.Close()
		return false, nil
	}
	var initPacketRet protocol.InitPacketRet
	netio.ReadPacket(conn, &initPacketRet)
	// 新建节点加入map
	serverNode := &Node{
		HashID:               utils.Array32ToUUID(initPacketRet.HashID),
		IsAdmin:              initPacketRet.IsAdmin,
		Conn:                 conn,
		ConnReadLock:         &sync.Mutex{},
		ConnWriteLock:        &sync.Mutex{},
		Socks5SessionIDLock:  &sync.Mutex{},
		Socks5DataBufferLock: &sync.RWMutex{},
		DirectConnection:     true,
	}
	Nodes[utils.Array32ToUUID(initPacketRet.HashID)] = serverNode

	serverNodeID := utils.Array32ToUUID(initPacketRet.HashID)
	GNetworkTopology.AddRoute(serverNodeID, serverNodeID)
	GNetworkTopology.AddNetworkMap(CurrentNode.HashID, serverNodeID)
	GNodeInfo.AddNode(serverNodeID)

	return true, serverNode
}
