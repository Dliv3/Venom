package dispather

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/Dliv3/Venom/global"
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
			node.Nodes[key] = &node.Node{
				HashID:               key,
				Conn:                 node.Nodes[value].Conn,
				ConnReadLock:         &sync.Mutex{},
				ConnWriteLock:        &sync.Mutex{},
				Socks5SessionIDLock:  &sync.Mutex{},
				Socks5DataBufferLock: &sync.RWMutex{},
			}
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
		DstHashID: utils.UUIDToArray32(global.CurrentPeerNodeHashID),
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
func SendConnectCmd(peerNode *node.Node, ip string, port uint16) {
	connectPacketCmd := protocol.ConnectPacketCmd{
		IP:   utils.IpToUint32(net.ParseIP(ip)),
		Port: port,
	}
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(global.CurrentPeerNodeHashID),
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
		DstHashID: utils.UUIDToArray32(global.CurrentPeerNodeHashID),
		CmdType:   protocol.DOWNLOAD,
	}
	peerNode.WritePacket(packetHeader, downloadPacketCmd)

	var retPacketHeader protocol.PacketHeader
	var downloadPacketRet protocol.DownloadPacketRet
	err = node.CurrentNode.CommandBuffers[protocol.DOWNLOAD].ReadPacket(&retPacketHeader, &downloadPacketRet)
	if err != nil {
		log.Println(fmt.Sprintf("[-]DownloadPacket Error: %s", err))
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
		log.Println(fmt.Sprintf("[-]DownloadPacket Error: %s", err))
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
