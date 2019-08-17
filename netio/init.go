package netio

import (
	"log"
	"net"

	"github.com/Dliv3/Venom/global"
)

// InitNode 初始化网络连接
// peerNodeID 存储需要通信(socks5/端口转发)的对端节点ID
func InitTCP(tcpType string, tcpService string, peerNodeID string, handlerFunc func(net.Conn, string, chan bool, ...interface{}), args ...interface{}) (err error) {
	if tcpType == "connect" {
		addr, err := net.ResolveTCPAddr("tcp", tcpService)
		if err != nil {
			log.Println("[-]ResolveTCPAddr error:", err)
			return err
		}

		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			log.Println("[-]DialTCP error:", err)
			return err
		}

		// conn.SetKeepAlive(true)

		go handlerFunc(conn, peerNodeID, nil, args)

		return err
	} else if tcpType == "listen" {
		var err error
		var listener net.Listener

		addr, err := net.ResolveTCPAddr("tcp", tcpService)
		if err != nil {
			log.Println("[-]ResolveTCPAddr error:", err)
			return err
		}
		listener, err = net.ListenTCP("tcp", addr)

		if err != nil {
			log.Println("[-]ListenTCP error:", err)
			return err
		}

		go func() {
			c := make(chan bool, global.TCP_MAX_CONNECTION)
			for {
				c <- true
				conn, err := listener.Accept()
				if err != nil {
					log.Println("[-]listener.Accept error:", err)
					// continue
					break
				}
				go handlerFunc(conn, peerNodeID, c, args)
			}
		}()
		return err
	}
	return INIT_TYPE_ERROR
}
