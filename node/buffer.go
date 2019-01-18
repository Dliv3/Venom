package node

import (
	"errors"
	"io"

	"github.com/Dliv3/Venom/protocol"
)

const DATA_BUFFER_SIZE = 4096

type Buffer struct {
	Chan chan interface{}
}

func NewBuffer() *Buffer {
	return &Buffer{
		Chan: make(chan interface{}, DATA_BUFFER_SIZE),
	}
}

func (buffer *Buffer) ReadLowLevelPacket() (protocol.Packet, error) {
	packet := <-buffer.Chan
	switch packet.(type) {
	case protocol.Packet:
		return packet.(protocol.Packet), nil
	// case error:
	// 	return protocol.Packet{}, io.EOF
	default:
		return protocol.Packet{}, errors.New("Data Type Error")
	}
}

func (buffer *Buffer) ReadPacket(packetHeader *protocol.PacketHeader, packetData interface{}) error {
	packet, err := buffer.ReadLowLevelPacket()
	if err != nil {
		return err
	}
	if packetHeader != nil {
		packet.ResolveHeader(packetHeader)
	}
	if packetData != nil {
		packet.ResolveData(packetData)
	}
	return nil
}

func (buffer *Buffer) WriteLowLevelPacket(packet protocol.Packet) {
	buffer.Chan <- packet
}

func (buffer *Buffer) WriteBytes(data []byte) {
	buffer.Chan <- data
}

func (buffer *Buffer) ReadBytes() ([]byte, error) {
	if buffer == nil {
		return nil, errors.New("Buffer is null")
	}
	data := <-buffer.Chan
	switch data.(type) {
	case []byte:
		return data.([]byte), nil
	// Fix Bug : socks5连接不会断开的问题
	case error:
		return nil, io.EOF
	default:
		return nil, errors.New("Data Type Error")
	}
}

// Fix Bug : socks5连接不会断开的问题
func (buffer *Buffer) WriteCloseMessage() {
	if buffer != nil {
		buffer.Chan <- io.EOF
	}
}
