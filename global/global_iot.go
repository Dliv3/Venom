// +build !386
// +build !amd64

package global

// 除去PacketHeader之外的部分最大允许使用的内存
// 防止单个数据包占用内存过大
const MAX_PACKET_SIZE = 128

// 一次最多可以承受16个连接
const TCP_MAX_CONNECTION = 16

// 命令&数据通道最大缓冲区大小
const BUFFER_SIZE = 64

// 协议数据分隔符
const PROTOCOL_SEPARATOR = "VCMD"

// 协议特征, 用于在端口重用时鉴别
const PROTOCOL_FEATURE = "ABCDEFGH"
