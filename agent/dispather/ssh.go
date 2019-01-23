package dispather

import (
	"fmt"
	"net"
	"time"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
	"golang.org/x/crypto/ssh"
)

const TIMEOUT = 5

func handleSshConnectCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var sshConnectPacketCmd protocol.SshConnectPacketCmd

		node.CurrentNode.CommandBuffers[protocol.SSHCONNECT].ReadPacket(&packetHeader, &sshConnectPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		var sshConnectPacketRet protocol.SshConnectPacketRet
		var packetHeaderRet protocol.PacketHeader

		packetHeaderRet = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SSHCONNECT,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}

		var auth ssh.AuthMethod
		if sshConnectPacketCmd.SshAuthMethod == 1 {
			auth = ssh.Password(string(sshConnectPacketCmd.SshAuthData))
		} else if sshConnectPacketCmd.SshAuthMethod == 2 {
			key, err := ssh.ParsePrivateKey(sshConnectPacketCmd.SshAuthData)
			if err != nil {
				sshConnectPacketRet.Success = 0
				sshConnectPacketRet.Msg = []byte("ssh key error")
				sshConnectPacketRet.MsgLen = uint32(len(sshConnectPacketRet.Msg))
				adminNode.WritePacket(packetHeaderRet, sshConnectPacketRet)
				continue
			}
			auth = ssh.PublicKeys(key)
		}

		config := ssh.ClientConfig{
			User: string(sshConnectPacketCmd.SshUser),
			Auth: []ssh.AuthMethod{
				auth,
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: time.Duration(time.Second * TIMEOUT),
		}

		sshClient, err := ssh.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", utils.Uint32ToIp(sshConnectPacketCmd.SshServer).String(), sshConnectPacketCmd.SshPort),
			&config,
		)

		if err != nil {
			sshConnectPacketRet.Success = 0
			sshConnectPacketRet.Msg = []byte(fmt.Sprintf("ssh connection error: %s", err))
			sshConnectPacketRet.MsgLen = uint32(len(sshConnectPacketRet.Msg))
			adminNode.WritePacket(packetHeaderRet, sshConnectPacketRet)
			continue
		}

		nodeConn, err := sshClient.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sshConnectPacketCmd.DstPort))

		if err != nil {
			sshConnectPacketRet.Success = 0
			sshConnectPacketRet.Msg = []byte(fmt.Sprintf("ssh connect to target node error: %s", err))
			sshConnectPacketRet.MsgLen = uint32(len(sshConnectPacketRet.Msg))
			adminNode.WritePacket(packetHeaderRet, sshConnectPacketRet)
			continue
		}

		AgentClient(nodeConn)

		sshConnectPacketRet.Success = 1
		sshConnectPacketRet.MsgLen = 0
		adminNode.WritePacket(packetHeaderRet, sshConnectPacketRet)
	}
}
