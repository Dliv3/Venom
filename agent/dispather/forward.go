package dispather

import (
	"io"
	"os/exec"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
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
