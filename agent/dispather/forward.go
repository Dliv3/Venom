package dispather

import (
	"io"
	"net"
	"os/exec"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
	"gitlab.com/Dliv3/Venom/util"
)

func CopyStdoutPipe2Node(input io.Reader, output *node.Node, c chan bool) {
	buf := make([]byte, global.MAX_PACKET_SIZE-8)
	for {
		count, err := input.Read(buf)
		data := protocol.ShellPacketRet{
			Success: 1,
			DataLen: uint32(count),
			Data:    buf[:count],
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
	}
	c <- true
	// fmt.Println("CopyStdoutPipe2Node Exit")

	return
}

func CopyNode2StdinPipe(input *node.Node, output io.Writer, c chan bool, cmd *exec.Cmd) {
	for {
		var packetHeader protocol.PacketHeader
		var shellPacketCmd protocol.ShellPacketCmd
		err := node.CurrentNode.CommandBuffers[protocol.SHELL].ReadPacket(&packetHeader, &shellPacketCmd)
		if shellPacketCmd.Start == 0 {
			break
		}
		if err != nil {
			break
		}
		output.Write(shellPacketCmd.Cmd)
		if string(shellPacketCmd.Cmd) == "exit\n" {
			break
		}
	}
	c <- true
	// fmt.Println("CopyNode2StdinPipe Exit")

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
		size, _ := util.PacketSize(socks5Data)
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SOCKSDATA,
			SrcHashID: util.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: util.UUIDToArray32(output.HashID),
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
		data, err := input.GetSocks5DataBuffer(currentSessionID).ReadBytes()
		if err != nil {
			c <- true
			break
		}
		output.Write(data)
	}
	c <- true
	return
}
