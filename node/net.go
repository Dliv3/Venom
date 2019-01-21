package node

import (
	"net"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

func CopyNet2Node(input net.Conn, output *Node, currentSessionID uint16, protocolType uint16, c chan bool) {
	buf := make([]byte, global.MAX_PACKET_SIZE-8)
	for {
		count, err := input.Read(buf)
		socks5Data := protocol.NetDataPacket{
			SessionID: currentSessionID,
			DataLen:   uint32(count),
			Data:      buf[:count],
		}
		size, _ := utils.PacketSize(socks5Data)
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocolType,
			SrcHashID: utils.UUIDToArray32(CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(output.HashID),
			DataLen:   size,
		}
		if err != nil {
			if count > 0 {
				output.WritePacket(packetHeader, socks5Data)
			}
			c <- true
			break
		}
		if count > 0 {
			output.WritePacket(packetHeader, socks5Data)
		}
	}
	c <- true
	return
}

func CopyNode2Net(input *Node, output net.Conn, currentSessionID uint16, protocolType uint16, c chan bool) {
	for {
		data, err := input.DataBuffers[protocolType].GetDataBuffer(currentSessionID).ReadBytes()
		if err != nil {
			c <- true
			break
		}
		output.Write(data)
	}
	c <- true
	return
}
