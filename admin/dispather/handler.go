package dispather

import (
	"fmt"
	"net"

	"github.com/Dliv3/Venom/node"
)

// AdminClient Admin节点作为Client
func AdminClient(conn net.Conn) {
	result, peerNode := node.ClentInitConnection(conn)
	if result {
		go node.CurrentNode.CommandHandler(peerNode)
	}
}

// AdminServer Admin节点作为Server
func AdminServer(conn net.Conn) {
	fmt.Println("\n[+]Remote connection: ", conn.RemoteAddr())
	result, peerNode := node.ServerInitConnection(conn)
	if result {
		fmt.Print("[+]A new node connect to admin node success")
		go node.CurrentNode.CommandHandler(peerNode)
	}
}
