package dispather

import (
	"bytes"
	"io"
	"runtime"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

func CopyStdin2Node(input io.Reader, output *node.Node, c chan bool) {

	orgBuf := make([]byte, global.MAX_PACKET_SIZE-8)

	for {
		count, err := input.Read(orgBuf)

		// fmt.Println(orgBuf[:count])

		var buf []byte

		// // delete \r
		if runtime.GOOS == "windows" {
			buf = bytes.Replace(orgBuf[:count], []byte("\r"), []byte(""), -1)
			count = len(buf)
		} else {
			buf = orgBuf
		}
		// fmt.Println(buf[:count])

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
