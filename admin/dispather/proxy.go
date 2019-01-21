package dispather

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/node"
	"github.com/Dliv3/Venom/protocol"
	"github.com/Dliv3/Venom/utils"
)

var (
	Commands = []string{"CONNECT", "BIND", "UDP ASSOCIATE"}
	AddrType = []string{"", "IPv4", "", "Domain", "IPv6"}
	Conns    = make([]net.Conn, 0)
	Verbose  = false

	errAddrType      = errors.New("socks addr type not supported")
	errVer           = errors.New("socks version not supported")
	errMethod        = errors.New("socks only support noauth method")
	errAuthExtraData = errors.New("socks authentication get extra data")
	errReqExtraData  = errors.New("socks request get extra data")
	errCmd           = errors.New("socks only support connect command")
)

const (
	socksVer5       = 0x05
	socksCmdConnect = 0x01
)

func AdminHandShake(conn net.Conn, peerNode *node.Node, currentSessionID uint16) (err error) {
	const (
		idVer     = 0
		idNmethod = 1
	)

	buf := make([]byte, 258)

	var n int

	// make sure we get the nmethod field
	// when node is admin

	if n, err = io.ReadAtLeast(conn, buf, idNmethod+1); err != nil {
		return
	}

	socks5Data := protocol.NetDataPacket{
		SessionID: currentSessionID,
		DataLen:   uint32(n),
		Data:      buf[:n],
	}
	size, _ := utils.PacketSize(socks5Data)
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		CmdType:   protocol.SOCKSDATA,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		DataLen:   size,
	}
	peerNode.WritePacket(packetHeader, socks5Data)

	if buf[idVer] != socksVer5 {
		return errVer
	}

	nmethod := int(buf[idNmethod]) //  client support auth mode
	msgLen := nmethod + 2          //  auth msg length
	if n == msgLen {               // handshake done, common case
		// do nothing, jump directly to send confirmation
	} else if n < msgLen { // has more methods to read, rare case
		if _, err = io.ReadFull(conn, buf[n:msgLen]); err != nil {
			return
		}
		socks5Data := protocol.NetDataPacket{
			SessionID: currentSessionID,
			DataLen:   uint32(msgLen - n),
			Data:      buf[n:msgLen],
		}
		size, _ := utils.PacketSize(socks5Data)
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SOCKSDATA,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(peerNode.HashID),
			DataLen:   size,
		}
		peerNode.WritePacket(packetHeader, socks5Data)
	} else { // error, should not get extra data
		return errAuthExtraData
	}
	return
}

func AdminParseTarget(conn net.Conn, peerNode *node.Node, currentSessionID uint16) (host string, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip addres start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4 = 1 // type is ipv4 address
		typeDm   = 3 // type is domain address
		typeIPv6 = 4 // type is ipv6 address

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)
	// refer to getRequest in server.go for why set buffer size to 263
	buf := make([]byte, 263)
	var n int

	// read till we get possible domain length field
	if n, err = io.ReadAtLeast(conn, buf, idDmLen+1); err != nil {
		return
	}

	socks5Data := protocol.NetDataPacket{
		SessionID: currentSessionID,
		DataLen:   uint32(n),
		Data:      buf[:n],
	}
	size, _ := utils.PacketSize(socks5Data)
	packetHeader := protocol.PacketHeader{
		Separator: global.PROTOCOL_SEPARATOR,
		CmdType:   protocol.SOCKSDATA,
		SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
		DstHashID: utils.UUIDToArray32(peerNode.HashID),
		DataLen:   size,
	}
	peerNode.WritePacket(packetHeader, socks5Data)

	// check version and cmd
	if buf[idVer] != socksVer5 {
		err = errVer
		return
	}

	/*
	   CONNECT X'01'
	   BIND X'02'
	   UDP ASSOCIATE X'03'
	*/

	if buf[idCmd] > 0x03 || buf[idCmd] == 0x00 {
		log.Println("Unknown Command", buf[idCmd])
	}

	if Verbose {
		log.Println("Command:", Commands[buf[idCmd]-1])
	}

	if buf[idCmd] != socksCmdConnect { //  only support CONNECT mode
		err = errCmd
		return
	}

	// read target address
	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm: // domain name
		reqLen = int(buf[idDmLen]) + lenDmBase
	default:
		err = errAddrType
		return
	}

	if n == reqLen {
		// common case, do nothing
	} else if n < reqLen { // rare case
		if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return
		}
		socks5Data := protocol.NetDataPacket{
			SessionID: currentSessionID,
			DataLen:   uint32(reqLen - n),
			Data:      buf[n:reqLen],
		}
		size, _ := utils.PacketSize(socks5Data)
		packetHeader := protocol.PacketHeader{
			Separator: global.PROTOCOL_SEPARATOR,
			CmdType:   protocol.SOCKSDATA,
			SrcHashID: utils.UUIDToArray32(node.CurrentNode.HashID),
			DstHashID: utils.UUIDToArray32(peerNode.HashID),
			DataLen:   size,
		}
		peerNode.WritePacket(packetHeader, socks5Data)
	} else {
		err = errReqExtraData
		return
	}

	switch buf[idType] {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}
	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return
}
