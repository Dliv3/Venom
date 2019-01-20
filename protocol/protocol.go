package protocol

import (
	"bytes"

	"github.com/Dliv3/Venom/netio"
)

// 为了方便自动化处理协议数据，只有分隔符可被设置成string类型
// 如果有某字段Abc使用了长度不确定的类型，如[]byte
// 则不需在该字段之前设置AbcLen字段来指定Abc字段长度

// 协议类型
const (
	// 初始化，在node对象建立之前
	INIT = iota
	// 控制协议
	SYNC
	LISTEN
	CONNECT
	SHELL
	UPLOAD
	DOWNLOAD
	SOCKS
	// 数据传输协议
	SOCKSDATA
	LFORWARDDATA
	RFORWARDDATA
)

// Packet 是较为低层的数据包格式
// 比Packet高层的网络数据存放在Data中
// 如果Packet不属于本节点，则直接转发即可，无需解析Data格式
type Packet struct {
	Separator string
	CmdType   uint16
	SrcHashID [32]byte // 源节点ID
	DstHashID [32]byte // 目的节点ID
	DataLen   uint64
	Data      []byte
}

// ResolveData 解析Packet Data字段的数据为特定格式的数据包
func (packet *Packet) ResolveData(cmdPacket interface{}) {
	netio.ReadPacket(bytes.NewBuffer(packet.Data), cmdPacket)
}

// ResolveHeader 解析Packet数据包中PacketHeader字段
func (packet *Packet) ResolveHeader(header *PacketHeader) {
	header.Separator = packet.Separator
	header.CmdType = packet.CmdType
	header.SrcHashID = packet.SrcHashID
	header.DstHashID = packet.DstHashID
	header.DataLen = packet.DataLen
}

type PacketHeader struct {
	Separator string
	CmdType   uint16
	SrcHashID [32]byte // 源节点ID
	DstHashID [32]byte // 目的节点ID
	DataLen   uint64
}

// InitPacketCmd 初始化数据包，命令数据
type InitPacketCmd struct {
	OsType  uint32 // 系统类型
	IsAdmin uint16 // 是否为管理员节点
	HashID  [32]byte
}

// InitPacketRet 初始化数据包，命令数据
type InitPacketRet struct {
	OsType  uint32 // 系统类型
	IsAdmin uint16 // 是否为管理员节点
	HashID  [32]byte
}

type SyncPacket struct {
	NetworkMapLen uint64
	NetworkMap    []byte
}

// UploadPacketCmd
type UploadPacketCmd struct {
	PathLen uint32 // 目标路径长度
	Path    []byte
	FileLen uint64 // 文件大小
	// File    []byte
}

// UploadPacketRet 文件上传返回包
type UploadPacketRet struct {
	Success uint16 // 操作是否成功， 1 or 0
	MsgLen  uint32 // 返回的信息长度
	Msg     []byte // 如果成功则为空, 否则为错误信息
}

// DownloadPacketCmd 文件下载命令
type DownloadPacketCmd struct {
	PathLen       uint32 // 目标路径长度
	Path          []byte // 目标路径名
	StillDownload uint32 // 如果文件过大是否还有继续下载
}

// DownloadPacketRet 文件下载返回包
type DownloadPacketRet struct {
	Success uint16 // 操作是否成功， 1 or 0
	MsgLen  uint32 // 返回的信息长度
	Msg     []byte // 如果成功则为空, 否则为错误信息
	FileLen uint64 // 文件大小
	// File    []byte
}

type FileDataPacket struct {
	DataLen uint32 // 返回的信息长度
	Data    []byte // 如果成功则为空, 否则为错误信息
}

type ListenPacketCmd struct {
	Port uint16
}

type ListenPacketRet struct {
	Success uint16 // 操作是否成功， 1 or 0
	MsgLen  uint32 // 返回的信息长度
	Msg     []byte // 如果成功则为空, 否则为错误信息
}

type ConnectPacketCmd struct {
	IP   uint32
	Port uint16
}

type ConnectPacketRet struct {
	Success uint16 // 操作是否成功， 1 or 0
	MsgLen  uint32 // 返回的信息长度
	Msg     []byte // 如果成功则为空, 否则为错误信息
}

type ShellPacketCmd struct {
	Start  uint16 // 启动shell
	CmdLen uint32 // 要执行的命令
	Cmd    []byte // 执行命令的长度
}

type ShellPacketRet struct {
	Success uint16 // 操作是否成功， 1 or 0
	DataLen uint32 // 返回的信息长度
	Data    []byte // 如果成功则为空, 否则为错误信息
}

type Socks5ControlPacketCmd struct {
	Start     uint16 // 启动一个socks5连接/关闭这个socks5连接，针对一个TCP连接而言
	SessionID uint16
}

type Socks5ControlPacketRet struct {
	Success uint16 // 启动一个socks5连接/关闭这个socks5连接，针对一个TCP连接而言
}

type Socks5DataPacket struct {
	SessionID uint16 // session id用于标识该数据包属于哪一个TCP连接，因为可能存在并发访问的问题
	// 在对数据包按命令类型分流之后，还需要对其中的socks5数据包做tcp会话分流
	DataLen uint32 // 返回的信息长度
	Data    []byte // 如果成功则为空, 否则为错误信息
	Close   uint16 // 1表示连接关闭
}

// type NetForwardPacketCmd struct {
// }

// type NetForwardPacketRet struct {
// }

// type NetForwardDataPacket struct {
// 	DataLen uint32
// 	Data    []byte
// }
