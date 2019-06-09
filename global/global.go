// +build 386 amd64

package global

// 除去PacketHeader之外的部分最大允许使用的内存
// 防止单个数据包占用内存过大
const MAX_PACKET_SIZE = 20480

// 一次最多可以承受7168个连接
const TCP_MAX_CONNECTION = 1024

// 命令&数据通道最大缓冲区大小
const BUFFER_SIZE = 1024

// 协议数据分隔符
var PROTOCOL_SEPARATOR = "VCMD"

// 协议特征, 用于在端口重用时鉴别
var PROTOCOL_FEATURE = "ABCDEFGH"

// 密钥
var SECRET_KEY []byte
