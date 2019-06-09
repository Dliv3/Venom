package protocol

import (
	"bytes"

	"github.com/Dliv3/Venom/crypto"
	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/netio"
)

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
	LFORWARD
	RFORWARD
	SSHCONNECT
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
	// 如果有解密需求, 先解密
	if global.SECRET_KEY != nil {
		// fmt.Println(packet.DataLen)
		// fmt.Println(packet.Data)
		packet.Data, _ = crypto.Decrypt(packet.Data, global.SECRET_KEY)
		packet.DataLen = uint64(len(packet.Data))
	}
	// fmt.Println(packet.Data)
	// fmt.Println(packet.DataLen)
	netio.ReadPacket(bytes.NewBuffer(packet.Data), cmdPacket)
}

// PackData 将cmdPacket打包成byte
func (packet *Packet) PackData(cmdPacket interface{}) {
	tmpBuffer := new(bytes.Buffer)
	netio.WritePacket(tmpBuffer, cmdPacket)
	packet.Data = tmpBuffer.Bytes()
	// 如果有加密需求, 后加密
	if global.SECRET_KEY != nil {
		packet.Data, _ = crypto.Encrypt(packet.Data, global.SECRET_KEY)
	}
	packet.DataLen = uint64(len(packet.Data))
}

// ResolveHeader 解析Packet数据包中PacketHeader字段
func (packet *Packet) ResolveHeader(header *PacketHeader) {
	header.Separator = packet.Separator
	header.CmdType = packet.CmdType
	header.SrcHashID = packet.SrcHashID
	header.DstHashID = packet.DstHashID
	header.DataLen = packet.DataLen
}

func (packet *Packet) PackHeader(header PacketHeader) {
	packet.Separator = header.Separator
	packet.CmdType = header.CmdType
	packet.SrcHashID = header.SrcHashID
	packet.DstHashID = header.DstHashID
	packet.DataLen = header.DataLen
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

type NetDataPacket struct {
	SessionID uint16 // session id用于标识该数据包属于哪一个TCP连接，因为可能存在并发访问的问题
	// 在对数据包按命令类型分流之后，还需要对其中的socks5数据包做tcp会话分流
	DataLen uint32 // 返回的信息长度
	Data    []byte // 如果成功则为空, 否则为错误信息
	Close   uint16 // 1表示连接关闭
}

type NetLForwardPacketCmd struct {
	Start   uint16
	DstPort uint16
	SrcPort uint16
	LHost   uint32
}

type NetLForwardPacketRet struct {
	Success   uint16
	SessionID uint16
	SrcPort   uint16
	LHost     uint32
}

type NetRForwardPacketCmd struct {
	Start     uint16
	SessionID uint16
	RHost     uint32
	SrcPort   uint16
}

type NetRForwardPacketRet struct {
	Success uint16
}

type SshConnectPacketCmd struct {
	SshServer      uint32 // 服务端IP地址
	SshPort        uint16 // ssh服务端口
	DstPort        uint16 // 要连接的target node监听的端口
	SshUserLen     uint32
	SshUser        []byte // ssh用户名
	SshAuthMethod  uint16 // password(1) / ssh key(2)
	SshAuthDataLen uint32
	SshAuthData    []byte // ssh key or password
}

type SshConnectPacketRet struct {
	Success uint16 // 操作是否成功， 1 or 0
	MsgLen  uint32 // 返回的信息长度
	Msg     []byte // 如果成功则为空, 否则为错误信息
}
