package node

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

// Node 节点
type Node struct {
	IsAdmin uint16   // Node是否是Admin
	HashID  string   // Node的HashID
	Conn    net.Conn // 与Node的TCP连接

	// Conn的锁，因为Conn读写Packet的时候如果不加锁，多个routine会出现乱序的情况
	ConnReadLock  *sync.Mutex
	ConnWriteLock *sync.Mutex

	// 控制信道缓冲区
	CommandBuffers map[uint16]*Buffer

	// 数据信道缓冲区
	DataBuffers map[uint16]*DataBuffer

	// 是否与本节点直接连接
	DirectConnection bool

	// Socks5Running bool // 防止admin node在一个agent上开启多个连接
}

func NewNode(isAdmin uint16, hashID string, conn net.Conn, directConnection bool) *Node {
	newNode := &Node{
		HashID:           hashID,
		IsAdmin:          isAdmin,
		Conn:             conn,
		ConnReadLock:     &sync.Mutex{},
		ConnWriteLock:    &sync.Mutex{},
		DirectConnection: directConnection,
	}
	newNode.InitDataBuffer()
	return newNode
}

// CommandHandler 协议数据包，将协议数据包分类写入Buffer
func (node *Node) CommandHandler(peerNode *Node) {
	defer peerNode.Disconnect()
	for {
		var lowLevelPacket protocol.Packet
		err := peerNode.ReadLowLevelPacket(&lowLevelPacket)
		if err != nil {
			fmt.Println("node disconnect: ", err)
			return
		}
		switch utils.Array32ToUUID(lowLevelPacket.DstHashID) {
		case node.HashID:
			if lowLevelPacket.Separator == global.PROTOCOL_SEPARATOR {
				switch lowLevelPacket.CmdType {
				case protocol.SYNC:
					fallthrough
				case protocol.LISTEN:
					fallthrough
				case protocol.CONNECT:
					fallthrough
				case protocol.SHELL:
					fallthrough
				case protocol.UPLOAD:
					fallthrough
				case protocol.DOWNLOAD:
					fallthrough
				case protocol.SOCKS:
					node.CommandBuffers[lowLevelPacket.CmdType].WriteLowLevelPacket(lowLevelPacket)
				case protocol.SOCKSDATA:
					var socks5Data protocol.Socks5DataPacket
					lowLevelPacket.ResolveData(&socks5Data)
					peerNodeID := utils.Array32ToUUID(lowLevelPacket.SrcHashID)
					if Nodes[peerNodeID].DataBuffers[protocol.SOCKSDATA].GetDataBuffer(socks5Data.SessionID) != nil {
						if socks5Data.Close == 1 {
							// Fix Bug : socks5连接不会断开的问题
							Nodes[peerNodeID].DataBuffers[protocol.SOCKSDATA].GetDataBuffer(socks5Data.SessionID).WriteCloseMessage()
						} else {
							// 只将数据写入数据buffer，不写入整个packet
							Nodes[peerNodeID].DataBuffers[protocol.SOCKSDATA].GetDataBuffer(socks5Data.SessionID).WriteBytes(socks5Data.Data)
						}
					}
				default:
					log.Println(fmt.Sprintf("[-]%s", ERR_UNKNOWN_CMD))
				}
			} else {
				log.Println("[-]Separator error")
			}
		default:
			// 如果节点为Agent节点转发
			if node.IsAdmin == 0 {
				nextNode := GNetworkTopology.RouteTable[utils.Array32ToUUID(lowLevelPacket.DstHashID)]
				targetNode := Nodes[nextNode]
				if targetNode != nil {
					targetNode.WriteLowLevelPacket(lowLevelPacket)
				} else {
					log.Println("[-]Can not find target node")
				}
			} else {
				// fmt.Println("src id:", utils.Array32ToUUID(lowLevelPacket.SrcHashID))
				// fmt.Println("dst id:", utils.Array32ToUUID(lowLevelPacket.DstHashID))
				// fmt.Println("dst cmd type:", lowLevelPacket.CmdType)
				fmt.Println("[-]Target node error")
			}
		}
	}
}

func (node *Node) InitCommandBuffer() {
	node.CommandBuffers = make(map[uint16]*Buffer)

	node.CommandBuffers[protocol.SYNC] = NewBuffer()
	node.CommandBuffers[protocol.LISTEN] = NewBuffer()
	node.CommandBuffers[protocol.CONNECT] = NewBuffer()
	node.CommandBuffers[protocol.SOCKS] = NewBuffer()
	node.CommandBuffers[protocol.UPLOAD] = NewBuffer()
	node.CommandBuffers[protocol.DOWNLOAD] = NewBuffer()
	node.CommandBuffers[protocol.SHELL] = NewBuffer()
}

func (node *Node) InitDataBuffer() {
	node.DataBuffers = make(map[uint16]*DataBuffer)

	node.DataBuffers[protocol.SOCKSDATA] = NewDataBuffer()
	node.DataBuffers[protocol.LFORWARDDATA] = NewDataBuffer()
	node.DataBuffers[protocol.RFORWARDDATA] = NewDataBuffer()
}

// TODO 只有与断掉节点之间相连的节点才会清理路由表/网络拓扑表/节点标号等
// 暂无法做到对全网所有节点的如下信息进行清理，这样有些麻烦，暂时也不是刚需
func (node *Node) Disconnect() {
	node.Conn.Close()
	// 删除网络拓扑
	GNetworkTopology.DeleteNode(node)
	// 删除节点
	delete(Nodes, node.HashID)
	// 删除结构体
	node = nil
}

func (node *Node) ReadLowLevelPacket(packet interface{}) error {
	node.ConnReadLock.Lock()
	defer node.ConnReadLock.Unlock()
	err := netio.ReadPacket(node.Conn, packet)
	if err != nil {
		return err
	}
	return nil
}

func (node *Node) WriteLowLevelPacket(packet interface{}) error {
	node.ConnWriteLock.Lock()
	defer node.ConnWriteLock.Unlock()
	err := netio.WritePacket(node.Conn, packet)
	if err != nil {
		return err
	}
	return nil
}

func (node *Node) ReadPacket(header *protocol.PacketHeader, packet interface{}) error {
	node.ConnReadLock.Lock()
	defer node.ConnReadLock.Unlock()

	// 读数据包的头部字段
	err := netio.ReadPacket(node.Conn, header)
	if err != nil {
		return err
	}
	// 读数据包的数据字段
	err = netio.ReadPacket(node.Conn, packet)
	if err != nil {
		return err
	}
	return nil
}

func (node *Node) WritePacket(header protocol.PacketHeader, packet interface{}) error {

	node.ConnWriteLock.Lock()
	defer node.ConnWriteLock.Unlock()

	// 写数据包的头部字段
	header.DataLen, _ = utils.PacketSize(packet)
	err := netio.WritePacket(node.Conn, header)
	if err != nil {
		return err
	}
	// 写数据包的数据字段
	err = netio.WritePacket(node.Conn, packet)
	if err != nil {
		return err
	}
	return nil
}

type NodeInfo struct {
	// 节点编号，已被分配的节点编号不会在节点断开后分给新加入网络的节点
	NodeNumber2UUID map[int]string
	NodeUUID2Number map[string]int
	// 节点描述
	NodeDescription map[int]string
}

// NodeExist 节点是否存在
func (info *NodeInfo) NodeExist(nodeID string) bool {
	if _, ok := info.NodeUUID2Number[nodeID]; ok {
		return true
	}
	return false
}

// AddNode 添加一个节点并为节点编号
func (info *NodeInfo) AddNode(nodeID string) {
	number := len(info.NodeNumber2UUID) + 1
	info.NodeNumber2UUID[number] = nodeID
	info.NodeUUID2Number[nodeID] = number
}

// UpdateNoteInfo 根据路由表信息给节点编号
func (info *NodeInfo) UpdateNoteInfo() {
	for key := range GNetworkTopology.RouteTable {
		if !info.NodeExist(key) {
			info.AddNode(key)
		}
	}
}
