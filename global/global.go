package global

// 除去PacketHeader之外的部分最大允许使用的内存
// 防止单个包占用内存过大
const MAX_PACKET_SIZE = 10240

// todo 需要在socks5客户端限制这个数量
// 一次最多可以承受7168个连接
// ulimit -n 最大65535
const SOCKS5_MAX_CONNECTION = 7168

// admin节点想要操作的对端节点的ID，主要用于goto命令
var CurrentPeerNodeHashID string

// 协议数据分隔符
const PROTOCOL_SEPARATOR = "TCMD"

// 协议特征, 用于在端口重用时鉴别
const PROTOCOL_FEATURE = "ABCDEFGH"
