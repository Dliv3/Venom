// +build !386
// +build !amd64

package netio

import (
	"errors"
	"fmt"
	"log"
	"net"
)

var INIT_TYPE_ERROR = errors.New("init type error")

func InitNode(tcpType string, tcpService string, handlerFunc func(net.Conn), portReuse bool, reusedPort uint16) (err error) {
	if portReuse {
		fmt.Println("iot device does not support port reuse.")
		return
	}
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

		conn.SetKeepAlive(true)

		go handlerFunc(conn)

		return nil
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
			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Println("[-]Accept error:", err)
					continue
				}

				go handlerFunc(conn)
			}
		}()
		return nil
	}
	return INIT_TYPE_ERROR
}
