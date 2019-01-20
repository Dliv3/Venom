package dispather

import (
	"io"
	"net"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

func CopyStdin2Node(input io.Reader, output *node.Node, c chan bool) {

	buf := make([]byte, global.MAX_PACKET_SIZE-8)

	for {
		count, err := input.Read(buf)
		data := protocol.ShellPacketCmd{
			Start:  1,
			CmdLen: uint32(count),
			Cmd:    buf[:count],
		}
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SHELL,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(output.HashID),
		}
		if err != nil {
			if count > 0 {
				output.WritePacket(packetHeader, data)
			}
			break
		}
		if count > 0 {
			output.WritePacket(packetHeader, data)
		}
		if string(buf[:count]) == "exit\n" {
			break
		}
	}
	c <- true
	// fmt.Println("CopyStdin2Node Exit")

	return
}

func CopyNode2Stdout(input *node.Node, output io.Writer, c chan bool) {
	for {
		var packetHeader protocol.PacketHeader
		var shellPacketRet protocol.ShellPacketRet
		err := node.CurrentNode.CommandBuffers[protocol.SHELL].ReadPacket(&packetHeader, &shellPacketRet)
		if err != nil {
			break
		}
		if shellPacketRet.Success == 0 {
			break
		}
		output.Write(shellPacketRet.Data)
	}
	c <- true
	// fmt.Println("CopyNode2Stdout Exit")

	return
}

func CopyNet2Node(input net.Conn, output *node.Node, currentSessionID uint16, c chan bool) {
	buf := make([]byte, global.MAX_PACKET_SIZE-8)
	for {
		count, err := input.Read(buf)
		socks5Data := protocol.Socks5DataPacket{
			SessionID: currentSessionID,
			DataLen:   uint32(count),
			Data:      buf[:count],
		}
		size, _ := utils.PacketSize(socks5Data)
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SOCKSDATA,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
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

func CopyNode2Net(input *node.Node, output net.Conn, currentSessionID uint16, c chan bool) {
	for {
		data, err := input.DataBuffers[protocol.SOCKSDATA].GetDataBuffer(currentSessionID).ReadBytes()
		if err != nil {
			c <- true
			break
		}
		output.Write(data)
	}
	c <- true
	return
}
