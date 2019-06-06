// +build !386
// +build !amd64

package netio

import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"strings"
	"github.com/Dliv3/Venom/myConn"
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

		conn_nocrypt, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			log.Println("[-]DialTCP error:", err)
			return err
		}

		conn_nocrypt.SetKeepAlive(true)

		conn, err := myConn.NewSecureConn(conn_nocrypt)
		if err != nil {
			log.Println("[-]NewSecureConn error:", err)
			return err
		}

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
				conn_nocrypt, err := listener.Accept()
				if err != nil {
					log.Println("[-]Accept error:", err)
					continue
				}

				/*  add by 00theway to encrypt net flows*/
				var conn net.Conn
				if strings.Index(runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name(),"localSocks5Server") != -1 {
					conn = conn_nocrypt

				}else {
					conn,err = myConn.NewSecureConn(conn_nocrypt)
					if err != nil {
						log.Println("[-]NewSecureConn error:", err)
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
