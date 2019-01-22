package node

import (
	"net"
	"runtime"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

func CopyNet2Node(input net.Conn, output *Node, currentSessionID uint16, protocolType uint16, c chan bool) error {
	buf := make([]byte, global.MAX_PACKET_SIZE-8)
	var err error
	var count int
	for {
		count, err = input.Read(buf)
		// fmt.Println(count)
		// fmt.Println(buf[:count])
		// fmt.Println(string(buf[:count]))
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
			// fmt.Println(err)
			if count > 0 {
				output.WritePacket(packetHeader, socks5Data)
			}
			break
		}
		if count > 0 {
			output.WritePacket(packetHeader, socks5Data)
		}
	}
	c <- true
	// fmt.Println("==================== CopyNet2Node Done! ====================")
	return err
}

func CopyNode2Net(input *Node, output net.Conn, currentSessionID uint16, protocolType uint16, c chan bool) error {
	var err error
	var data []byte
	for {
		data, err = input.DataBuffers[protocolType].GetDataBuffer(currentSessionID).ReadBytes()
		if err != nil {
			// fmt.Println(err)
			input.DataBuffers[protocolType].RealseDataBuffer(currentSessionID)
			runtime.GC()
			output.Close()
			break
		}
		output.Write(data)
	}
	// fmt.Println("==================== CopyNode2Net Done! ====================")
	c <- true
	return err
}
