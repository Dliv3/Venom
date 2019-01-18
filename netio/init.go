package netio

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Dliv3/Venom/global"
	reuseport "github.com/kavu/go_reuseport"
)

var INIT_TYPE_ERROR = errors.New("Init type error")

// func Init(tcpType string, tcpService string, handlerFunc interface{}, portReuse bool) (err error) {
func Init(tcpType string, tcpService string, handlerFunc func(net.Conn), portReuse bool) (err error) {
	if tcpType == "connect" {
		addr, err := net.ResolveTCPAddr("tcp", tcpService)
		if err != nil {
			log.Println("[-]ResolveTCPAddr Error:", err)
			return err
		}

		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			log.Println("[-]DialTCP Error:", err)
			return err
		}

		conn.SetKeepAlive(true)

		go handlerFunc(conn)

		return err
	} else if tcpType == "listen" {
		var err error
		var listener net.Listener

		if portReuse {
			listener, err = reuseport.Listen("tcp", tcpService)
		} else {
			addr, err := net.ResolveTCPAddr("tcp", tcpService)
			if err != nil {
				log.Println("[-]ResolveTCPAddr Error:", err)
				return err
			}
			listener, err = net.ListenTCP("tcp", addr)
		}

		if err != nil {
			log.Println("[-]ListenTCP Error:", err)
			return err
		}

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Println("[-]Accept Error:", err)
					continue
				}

				appProtocol, data := isAppProtocol(conn)
				if appProtocol {
					go func() {
						port := strings.Split(tcpService, ":")[1]
						addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%s", port))
						if err != nil {
							log.Println("[-]ResolveTCPAddr Error:", err)
							return
						}

						server, err := net.DialTCP("tcp", nil, addr)
						if err != nil {
							log.Println("[-]DialTCP Error:", err)
							return
						}

						Write(server, data)
						go NetCopy(conn, server)
						NetCopy(server, conn)
					}()
					continue
				}

				go handlerFunc(conn)
			}
		}()
		return err
	}
	return INIT_TYPE_ERROR
}

// isAppProtocol
// 返回值的第一个参数是标识协议是否为应用协议，判断前8字节是否为Venom发送的ABCDEFGH
// 如果不是则为应用协议，否则为Venom协议
func isAppProtocol(conn net.Conn) (bool, []byte) {
	var protocol = make([]byte, len(global.PROTOCOL_FEATURE))

	_, err := Read(conn, protocol)

	if err != nil {
		log.Println("[-]Read Protocol Packet Error: ", err)
		return false, protocol
	}

	if string(protocol) == global.PROTOCOL_FEATURE {
		return false, protocol
	} else {
		return true, protocol
	}
}
