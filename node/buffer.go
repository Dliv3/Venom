package node

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/protocol"
)

const TIME_OUT = 5

type Buffer struct {
	Chan chan interface{}
}

func NewBuffer() *Buffer {
	return &Buffer{
		Chan: make(chan interface{}, global.BUFFER_SIZE),
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
	// data := <-buffer.Chan
	select {
	case <-time.After(time.Second * TIME_OUT):
		return nil, errors.New("TimeOut")
	case data := <-buffer.Chan:
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
	// switch data.(type) {
	// case []byte:
	// 	return data.([]byte), nil
	// // Fix Bug : socks5连接不会断开的问题
	// case error:
	// 	return nil, io.EOF
	// default:
	// 	return nil, errors.New("Data Type Error")
	// }
}

// Fix Bug : socks5连接不会断开的问题
func (buffer *Buffer) WriteCloseMessage() {
	if buffer != nil {
		buffer.Chan <- io.EOF
	}
}

type DataBuffer struct {
	// 数据信道缓冲区
	DataBuffer     [global.TCP_MAX_CONNECTION]*Buffer
	DataBufferLock *sync.RWMutex

	// Session ID
	SessionID     uint16
	SessionIDLock *sync.Mutex
}

func NewDataBuffer() *DataBuffer {
	return &DataBuffer{
		SessionIDLock:  &sync.Mutex{},
		DataBufferLock: &sync.RWMutex{},
	}
}

func (dataBuffer *DataBuffer) GetDataBuffer(sessionID uint16) *Buffer {
	if int(sessionID) > len(dataBuffer.DataBuffer) {
		log.Println("[-]DataBuffer sessionID error: ", sessionID)
		return nil
	}
	dataBuffer.DataBufferLock.RLock()
	defer dataBuffer.DataBufferLock.RUnlock()
	return dataBuffer.DataBuffer[sessionID]
}

func (dataBuffer *DataBuffer) NewDataBuffer(sessionID uint16) {
	dataBuffer.DataBufferLock.Lock()
	defer dataBuffer.DataBufferLock.Unlock()
	dataBuffer.DataBuffer[sessionID] = NewBuffer()
}

func (dataBuffer *DataBuffer) RealseDataBuffer(sessionID uint16) {
	dataBuffer.DataBufferLock.Lock()
	defer dataBuffer.DataBufferLock.Unlock()
	dataBuffer.DataBuffer[sessionID] = nil
}

func (dataBuffer *DataBuffer) GetSessionID() uint16 {
	dataBuffer.SessionIDLock.Lock()
	defer dataBuffer.SessionIDLock.Unlock()
	id := dataBuffer.SessionID
	dataBuffer.SessionID = (dataBuffer.SessionID + 1) % global.TCP_MAX_CONNECTION
	return id
}
