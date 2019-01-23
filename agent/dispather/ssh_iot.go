// +build !386
// +build !amd64

package dispather

import (
	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

func handleSshConnectCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var sshConnectPacketCmd protocol.SshConnectPacketCmd

		node.CurrentNode.CommandBuffers[protocol.SSHCONNECT].ReadPacket(&packetHeader, &sshConnectPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		var sshConnectPacketRet protocol.SshConnectPacketRet

		packetHeaderRet := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SSHCONNECT,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}

		sshConnectPacketRet.Success = 0
		sshConnectPacketRet.Msg = []byte("iot device does not support ssh tunnel.")
		sshConnectPacketRet.MsgLen = uint32(len(sshConnectPacketRet.Msg))
		adminNode.WritePacket(packetHeaderRet, sshConnectPacketRet)
	}
}
