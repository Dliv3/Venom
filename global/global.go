package global

// 除去PacketHeader之外的部分最大允许使用的内存
// 防止单个数据包占用内存过大
const MAX_PACKET_SIZE = 10240

// 一次最多可以承受7168个连接
const TCP_MAX_CONNECTION = 7168

// 协议数据分隔符
const PROTOCOL_SEPARATOR = "VCMD"

// 协议特征, 用于在端口重用时鉴别
const PROTOCOL_FEATURE = "ABCDEFGH"
