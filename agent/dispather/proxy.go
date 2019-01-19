package dispather

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strconv"
	"time"

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

func AgentHandShake(peerNode *node.Node, currentSessionID uint16) (err error) {
	const (
		idVer     = 0
		idNmethod = 1
	)

	var n int

	buf, err := peerNode.GetSocks5DataBuffer(currentSessionID).ReadBytes()
	if err != nil {
		return err
	}

	n = len(buf)

	if buf[idVer] != socksVer5 {
		return errVer
	}

	nmethod := int(buf[idNmethod]) //  client support auth mode
	msgLen := nmethod + 2          //  auth msg length
	if n == msgLen {               // handshake done, common case
		// do nothing, jump directly to send confirmation
	} else if n < msgLen { // has more methods to read, rare case
		buf, err = peerNode.GetSocks5DataBuffer(currentSessionID).ReadBytes()
		if err != nil {
			return err
		}
	} else { // error, should not get extra data
		return errAuthExtraData
	}
	/*
	   X'00' NO AUTHENTICATION REQUIRED
	   X'01' GSSAPI
	   X'02' USERNAME/PASSWORD
	   X'03' to X'7F' IANA ASSIGNED
	   X'80' to X'FE' RESERVED FOR PRIVATE METHODS
	   X'FF' NO ACCEPTABLE METHODS
	*/
	// send confirmation: version 5, no authentication required
	// _, err = conn.Write([]byte{socksVer5, 0})
	buf = []byte{socksVer5, 0}
	socks5Data := protocol.Socks5DataPacket{
		SessionID: currentSessionID,
		DataLen:   uint32(len(buf)),
		Data:      buf,
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
	return
}

func AgentParseTarget(peerNode *node.Node, currentSessionID uint16) (host string, err error) {
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
	// buf := make([]byte, 263)
	var n int

	// read till we get possible domain length field
	// if n, err = io.ReadAtLeast(conn, buf, idDmLen+1); err != nil {
	// 	return
	// }

	buf, err := peerNode.GetSocks5DataBuffer(currentSessionID).ReadBytes()
	if err != nil {
		return
	}

	n = len(buf)

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
		// if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
		// 	return
		// }
		var tmp []byte
		tmp, err = peerNode.GetSocks5DataBuffer(currentSessionID).ReadBytes()
		if err != nil {
			return
		}
		buf = append(buf, tmp...)
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

func PipeWhenClose(peerNode *node.Node, currentSessionID uint16, target string) {

	// if Verbose {
	// log.Println("[+]Connect remote ", target, "...")
	// }

	// remoteConn, err := net.DialTimeout("tcp", target, time.Duration(time.Second*15))
	// change timeout to 3s
	// TODO 区分域名和IP两种情况，如果是IP的话，不要设置超时否则会导致nmap扫描很慢
	// 如果是域名可以设置超时，因为有域名解析时间
	remoteConn, err := net.DialTimeout("tcp", target, time.Duration(time.Second*2))
	// conn, err := net.DialTCP("tcp", nil, target)
	if err != nil {
		log.Println("Connect remote :", err)
		return
	}

	tcpAddr := remoteConn.LocalAddr().(*net.TCPAddr)
	if tcpAddr.Zone == "" {
		if tcpAddr.IP.Equal(tcpAddr.IP.To4()) {
			tcpAddr.Zone = "ip4"
		} else {
			tcpAddr.Zone = "ip6"
		}
	}

	// if Verbose {
	log.Println("[+]Connect to remote address:", target)
	// }

	rep := make([]byte, 256)
	rep[0] = 0x05
	rep[1] = 0x00 // success
	rep[2] = 0x00 //RSV

	//IP
	if tcpAddr.Zone == "ip6" {
		rep[3] = 0x04 //IPv6
	} else {
		rep[3] = 0x01 //IPv4
	}

	var ip net.IP
	if "ip6" == tcpAddr.Zone {
		ip = tcpAddr.IP.To16()
	} else {
		ip = tcpAddr.IP.To4()
	}
	pindex := 4
	for _, b := range ip {
		rep[pindex] = b
		pindex++
	}
	rep[pindex] = byte((tcpAddr.Port >> 8) & 0xff)
	rep[pindex+1] = byte(tcpAddr.Port & 0xff)
	// conn.Write(rep[0 : pindex+2])
	socks5Data := protocol.Socks5DataPacket{
		SessionID: currentSessionID,
		DataLen:   uint32(pindex + 2),
		Data:      rep[0 : pindex+2],
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

	// 退出处理
	defer remoteConn.Close()

	// Transfer data
	c := make(chan bool)

	// Copy local to remote
	go CopyNode2Net(peerNode, remoteConn, currentSessionID, c)

	// Copy remote to local
	go CopyNet2Node(remoteConn, peerNode, currentSessionID, c)

	<-c
}
