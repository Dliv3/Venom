// +build 386 amd64

package netio

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Dliv3/Venom/global"
	reuseport "github.com/libp2p/go-reuseport"
)

var INIT_TYPE_ERROR = errors.New("init type error")

const TIMEOUT = 5

// InitNode 初始化节点间网络连接
// handleFunc 处理net.Conn的函数
// portReuse 是否以端口重用的方式初始化网络连接
// reusedPort 被复用的端口，如果不使用端口复用，直接置零即可
func InitNode(tcpType string, tcpService string, handlerFunc func(net.Conn), portReuse bool, reusedPort uint16) (err error) {
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

		if portReuse {
			listener, err = reuseport.Listen("tcp", tcpService)
		} else {
			addr, err := net.ResolveTCPAddr("tcp", tcpService)
			if err != nil {
				log.Println("[-]ResolveTCPAddr error:", err)
				return err
			}
			listener, err = net.ListenTCP("tcp", addr)
		}

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

				if portReuse {
					appProtocol, data, timeout := isAppProtocol(conn)
					if appProtocol || (!appProtocol && timeout) {
						go func() {
							// port := strings.Split(tcpService, ":")[1]
							addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", reusedPort))
							if err != nil {
								log.Println("[-]ResolveTCPAddr error:", err)
								return
							}

							server, err := net.DialTCP("tcp", nil, addr)
							if err != nil {
								log.Println("[-]DialTCP error:", err)
								return
							}

							Write(server, data)
							go NetCopy(conn, server)
							NetCopy(server, conn)
						}()
						continue
					}
				}
				go handlerFunc(conn)
			}
		}()
		return nil
	}
	return INIT_TYPE_ERROR
}

// isAppProtocol
// 返回值的第一个参数是标识协议是否为应用协议，判断前8字节是否为Venom发送的ABCDEFGH
// 如果不是则为应用协议，否则为Venom协议
func isAppProtocol(conn net.Conn) (bool, []byte, bool) {
	var protocol = make([]byte, len(global.PROTOCOL_FEATURE))

	defer conn.SetReadDeadline(time.Time{})

	conn.SetReadDeadline(time.Now().Add(TIMEOUT * time.Second))

	count, err := Read(conn, protocol)

	timeout := false

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			timeout = true
			// mysql etc
			// fmt.Println("timeout")
			return false, protocol[:count], timeout
		} else {
			log.Println("[-]Read protocol packet error: ", err)
			return false, protocol[:count], timeout
		}
	}

	if string(protocol) == global.PROTOCOL_FEATURE {
		// is node
		return false, protocol[:count], timeout
	} else {
		// http etc
		return true, protocol[:count], timeout
	}
}
