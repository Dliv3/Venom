package node

import "github.com/Dliv3/Venom/utils"

// CurrentNode 当前节点
var CurrentNode = Node{
	IsAdmin:             0,
	HashID:              utils.NewUUID(),
	Conn:                nil, // 当前节点无需使用Conn
	ConnReadLock:        nil,
	ConnWriteLock:       nil,
	Socks5SessionIDLock: nil,
}

var GNetworkTopology = NetworkTopology{
	RouteTable: make(map[string]string),
	NetworkMap: make(map[string]([]string)),
}

var GNodeInfo = NodeInfo{
	NodeNumber2UUID: make(map[int]string),
	NodeUUID2Number: make(map[string]int),
	NodeDescription: make(map[int]string),
}

// Nodes 与当前节点连接的节点
var Nodes = make(map[string]*Node)
