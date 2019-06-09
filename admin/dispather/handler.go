package dispather

import (
	"fmt"
	"net"

	"github.com/Dliv3/Venom/crypto"
	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

// AdminClient Admin节点作为Client
func AdminClient(conn net.Conn) {
	result, peerNode := node.ClentInitConnection(conn)
	if result {
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

// AdminServer Admin节点作为Server
func AdminServer(conn net.Conn) {
	fmt.Println("\n[+]Remote connection: ", conn.RemoteAddr())
	result, peerNode := node.ServerInitConnection(conn)
	if result {
		fmt.Print("[+]A new node connect to admin node success")
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

var LForwardTarget map[string]bool

func InitAdminHandler() {
	LForwardTarget = make(map[string]bool)
	go handleLForward()
}

func handleLForward() {
	for {
		var packetHeaderRet protocol.PacketHeader
		var lforwardPacketRet protocol.NetLForwardPacketRet

		node.CurrentNode.CommandBuffers[protocol.LFORWARD].ReadPacket(&packetHeaderRet, &lforwardPacketRet)

		if lforwardPacketRet.Success == 0 {
			fmt.Println("lforward error on agent.")
			continue
		}

		peerNode := node.Nodes[utils.Array32ToUUID(packetHeaderRet.SrcHashID)]

		sessionID := lforwardPacketRet.SessionID

		// 初始化对应SessionID的Buffer
		peerNode.DataBuffers[protocol.LFORWARDDATA].NewDataBuffer(sessionID)

		lhost := utils.Uint32ToIp(lforwardPacketRet.LHost).String()
		sport := lforwardPacketRet.SrcPort

		if _, ok := LForwardTarget[crypto.Sha256(fmt.Sprintf("%s:%d", lhost, sport))]; !ok {
			continue
		}

		err := netio.InitTCP(
			"connect",
			fmt.Sprintf("%s:%d", lhost, sport),
			peerNode.HashID,
			func(conn net.Conn, peerNodeID string, done chan bool, args ...interface{}) {
				defer func() {
					// fmt.Println(" ################ admin close ################")
					closeData := protocol.NetDataPacket{
						SessionID: sessionID,
						Close:     1,
					}
					packetHeader := protocol.PacketHeader{
						Separator: global.PROTOCOL_SEPARATOR,
						CmdType:   protocol.LFORWARDDATA,
						SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
						DstHashID: utils.UUIDToArray32(peerNode.HashID),
					}
					peerNode.WritePacket(packetHeader, closeData)

					// peerNode.DataBuffers[protocol.LFORWARDDATA].RealseDataBuffer(sessionID)
					// runtime.GC()
				}()
				c := make(chan bool, 2)

				go node.CopyNet2Node(conn, peerNode, sessionID, protocol.LFORWARDDATA, c)
				go node.CopyNode2Net(peerNode, conn, sessionID, protocol.LFORWARDDATA, c)

				<-c
			})

		if err != nil {
			// fmt.Println("################ admin close ################")
			closeData := protocol.NetDataPacket{
				SessionID: sessionID,
				Close:     1,
			}
			packetHeader := protocol.PacketHeader{
				Separator: global.PROTOCOL_SEPARATOR,
				CmdType:   protocol.LFORWARDDATA,
				SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
				DstHashID: utils.UUIDToArray32(peerNode.HashID),
			}
			peerNode.WritePacket(packetHeader, closeData)

			// peerNode.DataBuffers[protocol.LFORWARDDATA].RealseDataBuffer(sessionID)
			// runtime.GC()
		}
	}
}
