package dispather

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

var ERR_UNKNOWN_CMD = errors.New("unknown command type")
var ERR_PROTOCOL_SEPARATOR = errors.New("unknown separator")
var ERR_TARGET_NODE = errors.New("can not find target node")
var ERR_FILE_EXISTS = errors.New("remote file already exists")
var ERR_FILE_NOT_EXISTS = errors.New("remote file not exists")

// AgentClient Admin节点作为Client
func AgentClient(conn net.Conn) {
	result, peerNode := node.ClentInitConnection(conn)
	if result {
		log.Println("[+]Connect to a new node success")
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

// AgentServer Admin节点作为Server
func AgentServer(conn net.Conn) {
	log.Println("[+]Remote connection: ", conn.RemoteAddr())
	result, peerNode := node.ServerInitConnection(conn)
	if result {
		log.Println("[+]A new node connect to this node success")
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

// InitAgentHandler Agent处理Admin发出的命令
func InitAgentHandler() {
	go handleSyncCmd()
	go handleListenCmd()
	go handleConnectCmd()
	go handleDownloadCmd()
	go handleUploadCmd()
	go handleShellCmd()
	go handleSocks5Cmd()
	go handleLForwardCmd()
	go handleRForwardCmd()
}

func handleSyncCmd() {
	for {
		// fmt.Println("Nodes", node.Nodes)

		var packetHeader protocol.PacketHeader
		var syncPacket protocol.SyncPacket
		node.CurrentNode.CommandBuffers[protocol.SYNC].ReadPacket(&packetHeader, &syncPacket)

		// 重新初始化网络拓扑，这样当有节点断开时网络拓扑会实时改变
		node.GNetworkTopology.InitNetworkMap()

		node.GNetworkTopology.ResolveNetworkMapData(syncPacket.NetworkMap)

		// 通信的对端节点
		var peerNodeID = utils.Array32ToUUID(packetHeader.SrcHashID)

		// nextNode为下一跳
		nextNode := node.Nodes[node.GNetworkTopology.RouteTable[peerNodeID]]

		// 递归向其他节点发送sync同步路由表请求
		for i := range node.Nodes {
			if node.Nodes[i].HashID != peerNodeID && node.Nodes[i].DirectConnection {
				tempPacketHeader := protocol.PacketHeader{
					Separator: global.PROTOCOL_SEPARATOR,
					SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
					DstHashID: utils.UUIDToArray32(node.Nodes[i].HashID),
					CmdType:   protocol.SYNC,
				}
				networkMap := node.GNetworkTopology.GenerateNetworkMapData()
				tempSyncPacket := protocol.SyncPacket{
					NetworkMapLen: uint64(len(networkMap)),
					NetworkMap:    networkMap,
				}

				node.Nodes[i].WritePacket(tempPacketHeader, tempSyncPacket)

				node.CurrentNode.CommandBuffers[protocol.SYNC].ReadPacket(&tempPacketHeader, &tempSyncPacket)

				node.GNetworkTopology.ResolveNetworkMapData(tempSyncPacket.NetworkMap)
			}
		}

		// 生成路由表
		node.GNetworkTopology.UpdateRouteTable()

		// fmt.Println("RouteTable", node.GNetworkTopology.RouteTable)

		// 创建Node结构体
		for key, value := range node.GNetworkTopology.RouteTable {
			if _, ok := node.Nodes[key]; !ok {
				// node.Nodes[key] = &node.Node{
				// 	HashID:        key,
				// 	Conn:          node.Nodes[value].Conn,
				// 	ConnReadLock:  &sync.Mutex{},
				// 	ConnWriteLock: &sync.Mutex{},
				// 	// Socks5SessionIDLock:  &sync.Mutex{},
				// 	// Socks5DataBufferLock: &sync.RWMutex{},
				// }
				// node.Nodes[key].InitDataBuffer()

				node.Nodes[key] = node.NewNode(
					0,
					key,
					node.Nodes[value].Conn,
					false,
				)
			}
		}

		// // 生成节点信息
		// node.GNodeInfo.UpdateNoteInfo()

		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
			CmdType:   protocol.SYNC,
		}
		networkMap := node.GNetworkTopology.GenerateNetworkMapData()
		syncPacket = protocol.SyncPacket{
			NetworkMapLen: uint64(len(networkMap)),
			NetworkMap:    networkMap,
		}
		nextNode.WritePacket(packetHeader, syncPacket)

		// fmt.Println(node.CurrentNode.HashID)
		// fmt.Println(node.GNetworkTopology.RouteTable)
		// fmt.Println(node.GNetworkTopology.NetworkMap)
	}
}

func handleListenCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var listenPacketCmd protocol.ListenPacketCmd
		node.CurrentNode.CommandBuffers[protocol.LISTEN].ReadPacket(&packetHeader, &listenPacketCmd)

		// adminNode := node.Nodes[node.GNetworkTopology.RouteTable[utils.Array32ToUUID(packetHeader.SrcHashID)]]

		// 网络拓扑同步完成之后即可直接使用以及构造好的节点结构体
		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		err := netio.InitNode(
			"listen",
			fmt.Sprintf("0.0.0.0:%d", listenPacketCmd.Port),
			AgentServer, false)

		var listenPacketRet protocol.ListenPacketRet
		if err != nil {
			listenPacketRet.Success = 0
			listenPacketRet.Msg = []byte(fmt.Sprintf("%s", err))
		} else {
			listenPacketRet.Success = 1
		}
		listenPacketRet.MsgLen = uint32(len(listenPacketRet.Msg))
		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.LISTEN,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeader, listenPacketRet)
	}
}

func handleConnectCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var connectPacketCmd protocol.ConnectPacketCmd

		node.CurrentNode.CommandBuffers[protocol.CONNECT].ReadPacket(&packetHeader, &connectPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		err := netio.InitNode(
			"connect",
			fmt.Sprintf("%s:%d", utils.Uint32ToIp(connectPacketCmd.IP).String(), connectPacketCmd.Port),
			AgentClient, false)

		var connectPacketRet protocol.ConnectPacketRet
		if err != nil {
			connectPacketRet.Success = 0
			connectPacketRet.Msg = []byte(fmt.Sprintf("%s", err))
		} else {
			connectPacketRet.Success = 1
		}
		connectPacketRet.MsgLen = uint32(len(connectPacketRet.Msg))
		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.CONNECT,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeader, connectPacketRet)
	}
}

func handleDownloadCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var downloadPacketCmd protocol.DownloadPacketCmd

		node.CurrentNode.CommandBuffers[protocol.DOWNLOAD].ReadPacket(&packetHeader, &downloadPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		filePath := string(downloadPacketCmd.Path)

		var downloadPacketRet protocol.DownloadPacketRet
		var file *os.File
		var fileSize int64
		// 如果文件存在，则下载
		if utils.FileExists(filePath) {
			var err error
			file, err = os.Open(filePath)
			if err != nil {
				downloadPacketRet.Success = 0
				downloadPacketRet.Msg = []byte(fmt.Sprintf("%s", err))
			} else {
				defer file.Close()
				downloadPacketRet.Success = 1
				fileSize = utils.GetFileSize(filePath)
				downloadPacketRet.FileLen = uint64(fileSize)
			}
		} else {
			downloadPacketRet.Success = 0
			downloadPacketRet.Msg = []byte(fmt.Sprintf("%s", ERR_FILE_NOT_EXISTS))
		}

		downloadPacketRet.MsgLen = uint32(len(downloadPacketRet.Msg))

		var retPacketHeader protocol.PacketHeader
		retPacketHeader.CmdType = protocol.DOWNLOAD
		retPacketHeader.Separator = global.PROTOCOL_SEPARATOR
		retPacketHeader.SrcHashID = packetHeader.DstHashID
		retPacketHeader.DstHashID = packetHeader.SrcHashID

		adminNode.WritePacket(retPacketHeader, downloadPacketRet)

		if downloadPacketRet.Success == 0 {
			continue
		}

		var cmdPacketHeader protocol.PacketHeader
		node.CurrentNode.CommandBuffers[protocol.DOWNLOAD].ReadPacket(&cmdPacketHeader, &downloadPacketCmd)

		if downloadPacketCmd.StillDownload == 0 {
			continue
		}

		adminNode.WritePacket(retPacketHeader, downloadPacketRet)

		var dataBlockSize = uint64(global.MAX_PACKET_SIZE - 4)
		loop := int64(downloadPacketRet.FileLen / dataBlockSize)
		remainder := downloadPacketRet.FileLen % dataBlockSize

		var size int64
		for ; loop >= 0; loop-- {
			var buf []byte
			if loop > 0 {
				buf = make([]byte, dataBlockSize)
			} else {
				buf = make([]byte, remainder)
			}
			n, err := io.ReadFull(file, buf)
			if n > 0 {
				size += int64(n)
				dataPacket := protocol.FileDataPacket{
					DataLen: uint32(n),
					Data:    buf[0:n],
				}
				retPacketHeader := protocol.PacketHeader{
					Separator: global.PROTOCOL_SEPARATOR,
					SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
					DstHashID: packetHeader.SrcHashID,
					CmdType:   protocol.DOWNLOAD,
				}
				adminNode.WritePacket(retPacketHeader, dataPacket)
			}
			if err != nil {
				if err != io.EOF {
					log.Println("[-]Read file error")
				}
				break
			}
		}
	}
}

func handleUploadCmd() {
	for {
		/* ------ before upload ------- */
		var packetHeader protocol.PacketHeader
		var uploadPacketCmd protocol.UploadPacketCmd
		node.CurrentNode.CommandBuffers[protocol.UPLOAD].ReadPacket(&packetHeader, &uploadPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		packetHeaderRet := protocol.PacketHeader{
			CmdType:   protocol.UPLOAD,
			Separator: global.PROTOCOL_SEPARATOR,
			SrcHashID: packetHeader.DstHashID,
			DstHashID: packetHeader.SrcHashID,
		}

		var uploadPacketRet protocol.UploadPacketRet

		var filePath = string(uploadPacketCmd.Path)

		var file *os.File
		// 如果文件不存在，则上传
		if !utils.FileExists(filePath) {
			var err error
			file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				uploadPacketRet.Success = 0
				uploadPacketRet.Msg = []byte(fmt.Sprintf("%s", err))
			} else {
				uploadPacketRet.Success = 1
				defer file.Close()
			}
		} else {
			uploadPacketRet.Success = 0
			uploadPacketRet.Msg = []byte(fmt.Sprintf("%s %s", filePath, ERR_FILE_EXISTS))
		}
		uploadPacketRet.MsgLen = uint32(len(uploadPacketRet.Msg))

		adminNode.WritePacket(packetHeaderRet, uploadPacketRet)

		if uploadPacketRet.Success == 0 || file == nil {
			continue
		}

		// /* ----- upload file -------- */
		node.CurrentNode.CommandBuffers[protocol.UPLOAD].ReadPacket(&packetHeader, &uploadPacketCmd)

		var uploadPacketRet2 protocol.UploadPacketRet

		var dataBlockSize = uint64(global.MAX_PACKET_SIZE - 4)
		loop := int64(uploadPacketCmd.FileLen / dataBlockSize)
		remainder := uploadPacketCmd.FileLen % dataBlockSize
		for loop >= 0 {
			if remainder != 0 {
				var fileDataPacket protocol.FileDataPacket
				var packetHeaderRet protocol.PacketHeader
				node.CurrentNode.CommandBuffers[protocol.UPLOAD].ReadPacket(&packetHeaderRet, &fileDataPacket)
				_, err := file.Write(fileDataPacket.Data)
				if err != nil {
					uploadPacketRet2.Success = 0
					uploadPacketRet2.Msg = []byte(fmt.Sprintf("%s", err))
				}
			}
			loop--
		}
		file.Close()

		uploadPacketRet2.Success = 1
		uploadPacketRet2.MsgLen = uint32(len(uploadPacketRet.Msg))
		adminNode.WritePacket(packetHeaderRet, uploadPacketRet2)
	}
}

func handleShellCmd() {

	for {

		var packetHeader protocol.PacketHeader
		var shellPacketCmd protocol.ShellPacketCmd

		node.CurrentNode.CommandBuffers[protocol.SHELL].ReadPacket(&packetHeader, &shellPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		if shellPacketCmd.Start != 1 {
			continue
		}

		var cmd *exec.Cmd

		switch utils.GetSystemType() {
		// windows
		case 0x01:
			cmd = exec.Command("c:\\windows\\system32\\cmd.exe")
		// mac , linux, others
		default:
			cmd = exec.Command("/bin/bash", "-i")
		}

		out, _ := cmd.StdoutPipe()
		in, _ := cmd.StdinPipe()
		cmd.Stderr = cmd.Stdout

		if err := cmd.Start(); err != nil {
			// log.Fatal(err)
			shellPacketRet := protocol.ShellPacketRet{
				Success: 0,
			}
			packetHeaderRet := protocol.PacketHeader{
				Separator: global.PROTOCOL_SEPARATOR,
				CmdType:   protocol.SHELL,
				SrcHashID: packetHeader.DstHashID,
				DstHashID: packetHeader.SrcHashID,
			}
			adminNode.WritePacket(packetHeaderRet, shellPacketRet)
			continue
		}

		shellPacketRet := protocol.ShellPacketRet{
			Success: 1,
		}
		packetHeaderRet := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SHELL,
			SrcHashID: packetHeader.DstHashID,
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeaderRet, shellPacketRet)

		c := make(chan bool, 2)
		go CopyNode2StdinPipe(adminNode, in, c, cmd)
		go CopyStdoutPipe2Node(out, adminNode, c)
		<-c
		<-c
		cmd.Wait()

		// exit
		ShellPacketRet := protocol.ShellPacketRet{
			Success: 0,
		}
		packetHeader = protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SHELL,
			SrcHashID: packetHeader.DstHashID,
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeader, ShellPacketRet)
	}
}

// handleSocks5Cmd 从node.CommandBuffers[protocol.SOCKS]中读取命令并处理
func handleSocks5Cmd() {
	for {
		// 启动socks5的命令数据包
		var packetHeader protocol.PacketHeader
		var socks5ControlCmd protocol.Socks5ControlPacketCmd
		node.CurrentNode.CommandBuffers[protocol.SOCKS].ReadPacket(&packetHeader, &socks5ControlCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		// 初始化对应SessionID的Buffer
		adminNode.DataBuffers[protocol.SOCKSDATA].NewDataBuffer(socks5ControlCmd.SessionID)

		// 返回启动成功命令
		socks5ControlRet := protocol.Socks5ControlPacketRet{
			Success: 1,
		}
		packetHeaderRet := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SOCKS,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: packetHeader.SrcHashID,
		}
		adminNode.WritePacket(packetHeaderRet, socks5ControlRet)

		go func() {
			defer func() {
				socks5CloseData := protocol.NetDataPacket{
					SessionID: socks5ControlCmd.SessionID,
					Close:     1,
				}
				packetHeader := protocol.PacketHeader{
					Separator: global.PROTOCOL_SEPARATOR,
					CmdType:   protocol.SOCKSDATA,
					SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
					DstHashID: utils.UUIDToArray32(adminNode.HashID),
				}
				adminNode.WritePacket(packetHeader, socks5CloseData)

				// adminNode.DataBuffers[protocol.SOCKSDATA].RealseDataBuffer(socks5ControlCmd.SessionID)
				// runtime.GC()
			}()
			if err := AgentHandShake(adminNode, socks5ControlCmd.SessionID); err != nil {
				log.Println("[-]Socks handshake error:", err)
				return
			}
			addr, err := AgentParseTarget(adminNode, socks5ControlCmd.SessionID)
			if err != nil {
				log.Println("[-]Socks consult transfer mode or parse target error:", err)
				return
			}
			PipeWhenClose(adminNode, socks5ControlCmd.SessionID, addr)
		}()
	}
}

func handleLForwardCmd() {
	for {
		var packetHeader protocol.PacketHeader
		var lforwardPacketCmd protocol.NetLForwardPacketCmd
		node.CurrentNode.CommandBuffers[protocol.LFORWARD].ReadPacket(&packetHeader, &lforwardPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		dport := lforwardPacketCmd.DstPort

		err := netio.InitTCP(
			"listen",
			fmt.Sprintf("0.0.0.0:%d", dport),
			adminNode.HashID,
			localLForwardServer,
			lforwardPacketCmd.LHost,
			lforwardPacketCmd.SrcPort,
		)

		lforwardPacketRet := protocol.NetLForwardPacketRet{
			SessionID: 0,
			Success:   1,
		}
		packetHeaderRet := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(adminNode.HashID),
			CmdType:   protocol.LFORWARD,
		}

		if err != nil {
			log.Println("[-]LForward tcp listen error")
			lforwardPacketRet.Success = 0
			node.Nodes[adminNode.HashID].WritePacket(packetHeaderRet, lforwardPacketRet)
		}
	}
}

func localLForwardServer(conn net.Conn, peerNodeID string, done chan bool, args ...interface{}) {
	// fmt.Println("localLForwardServer")
	// defer conn.Close()
	adminNode := node.Nodes[peerNodeID]
	currentSessionID := adminNode.DataBuffers[protocol.LFORWARDDATA].GetSessionID()

	defer func() {
		// fmt.Println("################ agent close ################")

		lforwardDataPacketCloseData := protocol.NetDataPacket{
			SessionID: currentSessionID,
			Close:     1,
		}
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.LFORWARDDATA,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(adminNode.HashID),
		}
		adminNode.WritePacket(packetHeader, lforwardDataPacketCloseData)

		// adminNode.DataBuffers[protocol.LFORWARDDATA].RealseDataBuffer(currentSessionID)
		// runtime.GC()
	}()

	adminNode.DataBuffers[protocol.LFORWARDDATA].NewDataBuffer(currentSessionID)

	lforwardPacketRet := protocol.NetLForwardPacketRet{
		SessionID: currentSessionID,
		Success:   1,
		LHost:     args[0].([]interface{})[0].(uint32),
		SrcPort:   args[0].([]interface{})[1].(uint16),
	}
	packetHeaderRet := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(adminNode.HashID),
		CmdType:   protocol.LFORWARD,
	}
	adminNode.WritePacket(packetHeaderRet, lforwardPacketRet)

	c := make(chan bool, 2)

	// 从node DataBuffer中读取数据，发送给客户端
	go node.CopyNet2Node(conn, adminNode, currentSessionID, protocol.LFORWARDDATA, c)
	go node.CopyNode2Net(adminNode, conn, currentSessionID, protocol.LFORWARDDATA, c)

	// exit
	<-c
	<-done
}

func handleRForwardCmd() {
	for {
		// 启动socks5的命令数据包
		var packetHeader protocol.PacketHeader
		var rforwardPacketCmd protocol.NetRForwardPacketCmd
		node.CurrentNode.CommandBuffers[protocol.RFORWARD].ReadPacket(&packetHeader, &rforwardPacketCmd)

		adminNode := node.Nodes[utils.Array32ToUUID(packetHeader.SrcHashID)]

		rhost := utils.Uint32ToIp(rforwardPacketCmd.RHost).String()
		sport := rforwardPacketCmd.SrcPort

		err := netio.InitTCP(
			"connect",
			fmt.Sprintf("%s:%d", rhost, sport),
			adminNode.HashID,
			func(conn net.Conn, peerNodeID string, done chan bool, args ...interface{}) {
				rforwardPacketRet := protocol.NetRForwardPacketRet{
					Success: 1,
				}
				packetHeaderRet := protocol.PacketHeader{
					Separator: global.PROTOCOL_SEPARATOR,
					CmdType:   protocol.RFORWARD,
					SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
					DstHashID: utils.UUIDToArray32(adminNode.HashID),
				}
				adminNode.WritePacket(packetHeaderRet, rforwardPacketRet)

				currentSessionID := rforwardPacketCmd.SessionID
				adminNode.DataBuffers[protocol.RFORWARDDATA].NewDataBuffer(currentSessionID)
				defer func() {
					closeData := protocol.NetDataPacket{
						SessionID: currentSessionID,
						Close:     1,
					}
					packetHeader := protocol.PacketHeader{
						Separator: global.PROTOCOL_SEPARATOR,
						CmdType:   protocol.RFORWARDDATA,
						SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
						DstHashID: utils.UUIDToArray32(adminNode.HashID),
					}
					adminNode.WritePacket(packetHeader, closeData)
				}()
				c := make(chan bool, 2)

				go node.CopyNode2Net(adminNode, conn, currentSessionID, protocol.RFORWARDDATA, c)
				go node.CopyNet2Node(conn, adminNode, currentSessionID, protocol.RFORWARDDATA, c)

				<-c
			})

		if err != nil {
			rforwardPacketRet := protocol.NetRForwardPacketRet{
				Success: 0,
			}
			packetHeaderRet := protocol.PacketHeader{
				Separator: global.PROTOCOL_SEPARATOR,
				CmdType:   protocol.RFORWARD,
				SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
				DstHashID: utils.UUIDToArray32(adminNode.HashID),
			}
			adminNode.WritePacket(packetHeaderRet, rforwardPacketRet)
		}
	}
}
