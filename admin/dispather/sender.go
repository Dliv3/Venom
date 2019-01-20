package dispather

import (
	"fmt"
	"io"
	"net"
	"os"
	"runtime"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
	"github.com/cheggaaa/pb"
)

// SendSyncCmd 发送同步网络拓扑的命令
func SendSyncCmd() {

	// 重新初始化网络拓扑，这样当有节点断开时网络拓扑会实时改变
	node.GNetworkTopology.InitNetworkMap()

	for i := range node.Nodes {
		// 向直连的节点发送SYNC数据包，同步网络拓扑
		// 目标节点会递归处理SYNC数据包，以获得全网拓扑
		if node.Nodes[i].DirectConnection {

			// 构造命令数据包
			packetHeader := protocol.PacketHeader{
				Separator: global.PROTOCOL_SEPARATOR,
				SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
				DstHashID: utils.UUIDToArray32(node.Nodes[i].HashID),
				CmdType:   protocol.SYNC,
			}
			// 生成本节点网络拓扑数据
			networkMap := node.GNetworkTopology.GenerateNetworkMapData()
			syncPacket := protocol.SyncPacket{
				NetworkMapLen: uint64(len(networkMap)),
				NetworkMap:    networkMap,
			}

			// 发送命令数据包
			node.Nodes[i].WritePacket(packetHeader, syncPacket)

			// 读取返回数据包
			node.CurrentNode.CommandBuffers[protocol.SYNC].ReadPacket(&packetHeader, &syncPacket)

			// 解析网络拓扑数据包
			node.GNetworkTopology.ResolveNetworkMapData(syncPacket.NetworkMap)
		}
	}

	// 生成路由表
	node.GNetworkTopology.UpdateRouteTable()

	// 生成节点信息
	node.GNodeInfo.UpdateNoteInfo()

	// fmt.Println(node.CurrentNode.HashID)
	// fmt.Println(node.GNetworkTopology.RouteTable)
	// fmt.Println(node.GNetworkTopology.NetworkMap)

	// 创建Node结构体
	// TODO 是否应该动态更新？目前觉得不需要，断掉的节点也可以留着，动态更新反而麻烦
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

}

// SendListenCmd 发送监听端口命令
func SendListenCmd(peerNode *node.Node, port uint16) {
	listenPacketCmd := protocol.ListenPacketCmd{
		Port: port,
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		CmdType:   protocol.LISTEN,
	}

	peerNode.WritePacket(packetHeader, listenPacketCmd)

	var listenPacketRet protocol.ListenPacketRet
	node.CurrentNode.CommandBuffers[protocol.LISTEN].ReadPacket(&packetHeader, &listenPacketRet)

	if listenPacketRet.Success == 1 {
		fmt.Println("listen local port success!")
	} else {
		fmt.Println("listen local port failed!")
		fmt.Println(string(listenPacketRet.Msg))
	}
}

// SendConnectCmd 发送连接命令
func SendConnectCmd(peerNode *node.Node, ip net.IP, port uint16) {
	connectPacketCmd := protocol.ConnectPacketCmd{
		IP:   utils.IpToUint32(ip),
		Port: port,
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		CmdType:   protocol.CONNECT,
	}

	peerNode.WritePacket(packetHeader, connectPacketCmd)

	var connectPacketRet protocol.ConnectPacketRet
	node.CurrentNode.CommandBuffers[protocol.CONNECT].ReadPacket(&packetHeader, &connectPacketRet)

	if connectPacketRet.Success == 1 {
		fmt.Println("connect to remote port success!")
	} else {
		fmt.Println("connect to remote port failed!")
		fmt.Println(string(connectPacketRet.Msg))
	}
}

// SendDownloadCmd 发送下载命令
func SendDownloadCmd(peerNode *node.Node, remotePath string, localPath string) bool {
	/* ----------- before download file ---------- */
	if utils.FileExists(localPath) {
		fmt.Println("local file already exists")
		return false
	}

	localFile, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer localFile.Close()

	downloadPacketCmd := protocol.DownloadPacketCmd{
		PathLen: uint32(len(remotePath)),
		Path:    []byte(remotePath),
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		CmdType:   protocol.DOWNLOAD,
	}
	peerNode.WritePacket(packetHeader, downloadPacketCmd)

	var retPacketHeader protocol.PacketHeader
	var downloadPacketRet protocol.DownloadPacketRet
	err = node.CurrentNode.CommandBuffers[protocol.DOWNLOAD].ReadPacket(&retPacketHeader, &downloadPacketRet)
	if err != nil {
		fmt.Println(fmt.Sprintf("downloadpacket error: %s", err))
		return false
	}

	if downloadPacketRet.Success == 0 {
		fmt.Println("download file error: " + string(downloadPacketRet.Msg))
		if utils.FileExists(localPath) {
			os.Remove(localPath)
		}
		return false
	}

	if downloadPacketRet.FileLen > 1024*1024*100 {
		fmt.Print("this file is too large(>100M), still downloading? (y/n)")
		var choise string
		fmt.Scanf("%s", &choise)
		if choise != "y" {
			fmt.Println("stop download.")
			downloadPacketCmd.StillDownload = 0
			peerNode.WritePacket(packetHeader, downloadPacketCmd)

			if utils.FileExists(localPath) {
				os.Remove(localPath)
			}
			return false
		}
	}

	downloadPacketCmd.StillDownload = 1
	peerNode.WritePacket(packetHeader, downloadPacketCmd)

	/* ---------- download file ---------- */
	err = node.CurrentNode.CommandBuffers[protocol.DOWNLOAD].ReadPacket(&packetHeader, &downloadPacketRet)
	if err != nil {
		fmt.Println(fmt.Sprintf("[-]downloadpacket error: %s", err))
		return false
	}

	// 开始下载文件
	var dataBlockSize = uint64(global.MAX_PACKET_SIZE - 4)
	loop := int64(downloadPacketRet.FileLen / dataBlockSize)
	remainder := downloadPacketRet.FileLen % dataBlockSize

	// 进度条功能
	bar := pb.New64(int64(downloadPacketRet.FileLen))

	// show percents (by default already true)
	bar.ShowPercent = true

	// show bar (by default already true)
	bar.ShowBar = true

	bar.ShowCounters = true

	bar.ShowTimeLeft = true

	bar.SetUnits(pb.U_BYTES)

	// and start
	bar.Start()

	for ; loop >= 0; loop-- {
		if remainder != 0 {
			var fileDataPacket protocol.FileDataPacket
			node.CurrentNode.CommandBuffers[protocol.DOWNLOAD].ReadPacket(&packetHeader, &fileDataPacket)
			_, err = localFile.Write(fileDataPacket.Data)
			if err != nil {
				fmt.Println(err)
			}
			bar.Add64(int64(fileDataPacket.DataLen))
		}
	}
	bar.Finish()
	fmt.Println("download file success!")

	return true
}

// SendUploadCmd 发送上传命令
func SendUploadCmd(peerNode *node.Node, localPath string, remotePath string) bool {
	if !utils.FileExists(localPath) {
		fmt.Println("local file does not exists")
		return false
	}
	localFile, err := os.Open(localPath)
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer localFile.Close()

	// 如果文件过大，提醒用户选择是否继续上次（过大的文件会影响其他命令数据的传输效率）
	var fileSize = utils.GetFileSize(localPath)
	if fileSize > 1024*1024*100 {
		fmt.Print("this file is too large(>100M), still uploading? (y/n)")
		var choise string
		fmt.Scanf("%s", &choise)
		if choise != "y" {
			fmt.Println("stop upload.")
			return false
		}
	}

	/* ----- before upload ----- */
	// 在文件上传前，首先要确定remotePath没有错误
	uploadPacketCmd := protocol.UploadPacketCmd{
		PathLen: uint32(len(remotePath)),
		Path:    []byte(remotePath),
		FileLen: uint64(fileSize),
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		CmdType:   protocol.UPLOAD,
	}

	peerNode.WritePacket(packetHeader, uploadPacketCmd)

	var packetHeaderRet protocol.PacketHeader
	var uploadPacketRet protocol.UploadPacketRet
	err = node.CurrentNode.CommandBuffers[protocol.UPLOAD].ReadPacket(&packetHeaderRet, &uploadPacketRet)
	if err != nil {
		fmt.Println(fmt.Sprintf("syncpacket error: %s", err))
		return false
	}
	if uploadPacketRet.Success == 0 {
		fmt.Println("upload file error: " + string(uploadPacketRet.Msg))
		return false
	}
	/* ----- upload file ------- */
	peerNode.WritePacket(packetHeader, uploadPacketCmd)

	// 单个数据包最大为MAX_PACKET_SIZE，除去非数据字段DataLen占用4字节
	var dataBlockSize = int64(global.MAX_PACKET_SIZE - 4)
	loop := fileSize / dataBlockSize
	remainder := fileSize % dataBlockSize

	// 进度条功能
	bar := pb.New64(fileSize)

	// show percents (by default already true)
	bar.ShowPercent = true

	// show bar (by default already true)
	bar.ShowBar = true

	bar.ShowCounters = true

	bar.ShowTimeLeft = true

	bar.SetUnits(pb.U_BYTES)

	// and start
	bar.Start()

	var size int64
	// TODO: 直接在文件协议数据包中写明会传输几个数据包，而不要使用loop决定
	for ; loop >= 0; loop-- {
		var buf []byte
		if loop > 0 {
			buf = make([]byte, dataBlockSize)
		} else {
			buf = make([]byte, remainder)
		}
		// n, err := localFile.Read(buf[0:])
		n, err := io.ReadFull(localFile, buf)
		if n > 0 {
			size += int64(n)
			dataPacket := protocol.FileDataPacket{
				DataLen: uint32(n),
				Data:    buf[0:n],
			}
			dataLen, _ := utils.PacketSize(dataPacket)
			packetHeader := protocol.PacketHeader{
				Separator: global.PROTOCOL_SEPARATOR,
				SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
				DstHashID: utils.UUIDToArray32(peerNode.HashID),
				CmdType:   protocol.UPLOAD,
				DataLen:   dataLen,
			}
			peerNode.WritePacket(packetHeader, dataPacket)
			bar.Add64(int64(dataLen))
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println("[-]read file error")
			}
			break
		}
	}
	bar.Finish()

	err = node.CurrentNode.CommandBuffers[protocol.UPLOAD].ReadPacket(&packetHeaderRet, &uploadPacketRet)

	if err != nil {
		fmt.Println(fmt.Sprintf("[-]syncpacket error: %s", err))
		return false
	}
	if uploadPacketRet.Success == 0 {
		fmt.Println("upload file error: " + string(uploadPacketRet.Msg))
		return false
	}
	fmt.Println("upload file success!")
	return true
}

// SendShellCmd 发送shell命令
func SendShellCmd(peerNode *node.Node) {

	shellPacketCmd := protocol.ShellPacketCmd{
		Start: 1,
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		CmdType:   protocol.SHELL,
	}

	peerNode.WritePacket(packetHeader, shellPacketCmd)

	var packetHeaderRet protocol.PacketHeader
	var shellPacketRet protocol.ShellPacketRet
	node.CurrentNode.CommandBuffers[protocol.SHELL].ReadPacket(&packetHeaderRet, &shellPacketRet)

	if shellPacketRet.Success == 1 {
		c := make(chan bool, 2)
		go CopyStdin2Node(os.Stdin, peerNode, c)
		go CopyNode2Stdout(peerNode, os.Stdout, c)
		<-c
		<-c
		// exit = true
	} else {
		fmt.Println("something error.")
	}
}

// SendSocks5Cmd 启动socks5代理
func SendSocks5Cmd(peerNode *node.Node, port uint16) bool {
	err := netio.InitTCP("listen", fmt.Sprintf("0.0.0.0:%d", port), localSocks5Server, peerNode.HashID)

	if err != nil {
		fmt.Println("socks5 proxy startup error")
		return false
	}
	fmt.Printf("a socks5 proxy of the target node has started up on local port %d\n", port)
	return true
}

func localSocks5Server(conn net.Conn, peerNodeID string, done chan bool) {
	defer conn.Close()

	peerNode := node.Nodes[peerNodeID]

	currentSessionID := node.Nodes[peerNodeID].DataBuffers[protocol.SOCKSDATA].GetSessionID()

	defer func() {
		// Fix Bug : socks5连接不会断开的问题
		socks5CloseData := protocol.Socks5DataPacket{
			SessionID: currentSessionID,
			Close:     1,
		}
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SOCKSDATA,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(peerNode.HashID),
		}
		peerNode.WritePacket(packetHeader, socks5CloseData)

		node.Nodes[peerNodeID].DataBuffers[protocol.SOCKSDATA].RealseDataBuffer(currentSessionID)
		runtime.GC()
	}()

	socks5ControlCmd := protocol.Socks5ControlPacketCmd{
		SessionID: currentSessionID,
		Start:     1,
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNodeID),
		CmdType:   protocol.SOCKS,
	}
	// send socks5 start command, send session id to socks5 server node
	node.Nodes[peerNodeID].WritePacket(packetHeader, socks5ControlCmd)

	// ReadPacket From CommandBuffer
	var packetHeaderRet protocol.PacketHeader
	var socks5ControlRet protocol.Socks5ControlPacketRet
	node.CurrentNode.CommandBuffers[protocol.SOCKS].ReadPacket(&packetHeaderRet, &socks5ControlRet)

	if socks5ControlRet.Success == 0 {
		fmt.Println("socks5 start error On agent")
		return
	}

	// start read socks5 data from socks5 client
	// socks5 data buffer
	node.Nodes[peerNodeID].DataBuffers[protocol.SOCKSDATA].NewDataBuffer(currentSessionID)

	c := make(chan bool)

	// 从node Socks5Buffer中读取数据，发送给客户端
	go CopyNode2Net(peerNode, conn, currentSessionID, c)

	if err := AdminHandShake(conn, peerNode, currentSessionID); err != nil {
		fmt.Println("socks handshake:")
		fmt.Println(err)
		return
	}
	_, err := AdminParseTarget(conn, peerNode, currentSessionID)
	if err != nil {
		fmt.Println("socks consult transfer mode or parse target :")
		fmt.Println(err)
		return
	}

	// 从本地socket接收数据，发送给服务端
	go CopyNet2Node(conn, peerNode, currentSessionID, c)

	// exit
	<-c
	<-done
}
