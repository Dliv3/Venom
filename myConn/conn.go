package myConn

/*  add by 00theway to encrypt net flows*/

import (
	"bytes"
	"encoding/binary"
	"net"
)

const (
	MsgLenFieldSize = 4
)

type secureConn struct {
	net.Conn
	crypt    AesCrypt
	overhead int
	input    bytes.Reader
}


func (p * secureConn) ReciveMsgbuf() (int,error){
	var msgSize uint32
	var decrypted []byte
	var msgSizeBuf []byte
	var msgBuf []byte

	msgSizeBuf = make([]byte, p.overhead)
	_, err := p.Conn.Read(msgSizeBuf)
	if err != nil {
		return 0, err
	}
	msgSize = binary.LittleEndian.Uint32(msgSizeBuf)

	msgBuf = make([]byte, msgSize)
	_, err = p.Conn.Read(msgBuf)
	if err != nil {
		return 0, err
	}

	if msgSize != 0 {
		decrypted, err = p.crypt.Decrypt(decrypted, msgBuf)
		if err != nil {
			return 0, err
		}
	}
	p.input.Reset(decrypted)
	return 0, nil
}

func (p *secureConn) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	for p.input.Len() == 0 {
		_,err := p.ReciveMsgbuf()
		if err != nil {
			return 0, err
		}
	}

	n, _ := p.input.Read(buf)
	return n, nil
}

func (p *secureConn) Write(rawBuf []byte) (int, error) {
	var buf []byte

	buf, err := p.crypt.Encrypt(buf, rawBuf)
	if err != nil {
		return 0, err
	}
	msg := make([]byte, len(buf)+p.overhead)

	copy(msg[4:], buf)
	msgSize := uint32(len(msg) - p.overhead)
	binary.LittleEndian.PutUint32(msg, msgSize)
	_, err = p.Conn.Write(msg)
	if err != nil {
		return 0, err
	}
	return len(rawBuf), nil
}

func NewSecureConn(c net.Conn) (net.Conn, error) {
	crypt := &AesCrypt{}
	overhead := MsgLenFieldSize
	helloConn := &secureConn{
		Conn:     c,
		crypt:    *crypt,
		overhead: overhead,
	}

	return helloConn, nil
}