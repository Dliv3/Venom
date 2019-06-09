package netio

import (
	"errors"
	"io"
	"log"
	"net"
	"reflect"

	"github.com/Dliv3/Venom/crypto"
	"github.com/Dliv3/Venom/global"
	"github.com/Dliv3/Venom/utils"
)

// WritePacket write packet to node.Conn
func WritePacket(output io.Writer, packet interface{}) error {
	t := reflect.TypeOf(packet)
	v := reflect.ValueOf(packet)

	if k := t.Kind(); k != reflect.Struct {
		return errors.New("second param is not struct")
	}

	count := t.NumField()
	for i := 0; i < count; i++ {
		val := v.Field(i).Interface()

		// type switch
		switch value := val.(type) {
		case uint16:
			_, err := Write(output, utils.Uint16ToBytes(value))
			if err != nil {
				return err
			}
		case uint32:
			_, err := Write(output, utils.Uint32ToBytes(value))
			if err != nil {
				return err
			}
		case uint64:
			_, err := Write(output, utils.Uint64ToBytes(value))
			if err != nil {
				return err
			}
		case string:
			_, err := Write(output, []byte(value))
			if err != nil {
				return err
			}
		case []byte:
			_, err := Write(output, value)
			if err != nil {
				return err
			}
		case [2]byte:
			_, err := Write(output, value[0:])
			if err != nil {
				return err
			}
		case [4]byte:
			_, err := Write(output, value[0:])
			if err != nil {
				return err
			}
		case [32]byte:
			_, err := Write(output, value[0:])
			if err != nil {
				return err
			}
		default:
			return errors.New("type unsupport")
		}
	}
	return nil
}

// ReadPacket read packet from node.Conn
// packet data start from the packet separator
func ReadPacket(input io.Reader, packet interface{}) error {
	v := reflect.ValueOf(packet)
	t := reflect.TypeOf(packet)

	if v.Kind() == reflect.Ptr && !v.Elem().CanSet() {
		return errors.New("packet is not a reflect. Ptr or elem can not be setted")
	}

	v = v.Elem()

	t = t.Elem()
	count := t.NumField()

	for i := 0; i < count; i++ {
		val := v.Field(i).Interface()
		f := v.FieldByName(t.Field(i).Name)

		// 类型断言
		switch val.(type) {
		case string:
			// 字段为分隔符，只有分隔符字段可被设置成string类型
			// 在处理协议数据包之前，首先读取到协议数据分隔符
			// 分隔符为协议结构体的第一个数据
			if i == 0 {
				separator, err := readUntilSeparator(input, global.PROTOCOL_SEPARATOR)
				if err != nil {
					return err
				}
				f.SetString(separator)
			}
		case uint16:
			var buf [2]byte
			_, err := Read(input, buf[0:])
			if err != nil {
				return err
			}
			f.SetUint(uint64(utils.BytesToUint16(buf[0:])))
		case uint32:
			var buf [4]byte
			_, err := Read(input, buf[0:])
			if err != nil {
				return err
			}
			f.SetUint(uint64(utils.BytesToUint32(buf[0:])))
		case uint64:
			var buf [8]byte
			_, err := Read(input, buf[0:])
			if err != nil {
				return err
			}
			f.SetUint(uint64(utils.BytesToUint64(buf[0:])))
		case []byte:
			// 要求, 未指明长度的字段名需要有字段来指定其长度，并长度字段名为该字段名+Len
			// 如HashID字段是通过HashIDLen指明长度的
			// 并且要求HashIDLen在结构体中的位置在HashID之前
			temp := v.FieldByName(t.Field(i).Name + "Len")
			// 类型断言，要求长度字段类型必须为uint16、uint32或uint64
			var length uint64
			switch lengthTemp := temp.Interface().(type) {
			case uint64:
				length = lengthTemp
			case uint32:
				length = uint64(lengthTemp)
			case uint16:
				length = uint64(lengthTemp)
			}
			// 如果长度为0，就不需要读数据了
			if length != 0 {
				if length > global.MAX_PACKET_SIZE+uint64(crypto.OVERHEAD) {
					return nil
				}
				buf := make([]byte, length)
				_, err := Read(input, buf[0:])
				if err != nil {
					return err
				}
				f.SetBytes(buf)
			}
		case [2]byte:
			var buf [2]byte
			_, err := Read(input, buf[0:])
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(buf))
		case [4]byte:
			var buf [4]byte
			_, err := Read(input, buf[0:])
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(buf))
		case [32]byte:
			var buf [32]byte
			_, err := Read(input, buf[0:])
			if err != nil {
				return err
			}
			// 使用reflect给array类型赋值的方法
			f.Set(reflect.ValueOf(buf))
		default:
			return errors.New("type unsupport")
		}
	}
	return nil
}

func Read(input io.Reader, buffer []byte) (int, error) {
	n, err := io.ReadFull(input, buffer)
	if err != nil {
		// log.Println("[-]Read Error: ", err)
	}
	return n, err
}

func Write(output io.Writer, buffer []byte) (int, error) {
	if len(buffer) > 0 {
		n, err := output.Write(buffer)
		if err != nil {
			// log.Println("[-]Write Error: ", err)
		}
		return n, err
	}
	return 0, nil
}

// if found, return PROTOCOL_SEPARATOR
func readUntilSeparator(input io.Reader, separator string) (string, error) {
	kmp, _ := utils.NewKMP(separator)
	i := 0
	var one [1]byte
	for {
		_, err := Read(input, one[0:])
		if err != nil {
			return "", err
		}
		if kmp.Pattern[i] == one[0] {
			if i == kmp.Size-1 {
				return kmp.Pattern, nil
			}
			i++
			continue
		}
		if kmp.Prefix[i] > -1 {
			i = kmp.Prefix[i]
		} else {
			i = 0
		}
	}
}

func NetCopy(input, output net.Conn) (err error) {
	defer input.Close()

	buf := make([]byte, global.MAX_PACKET_SIZE)
	for {
		count, err := input.Read(buf)
		if err != nil {
			if err == io.EOF && count > 0 {
				output.Write(buf[:count])
			}
			if err != io.EOF {
				log.Fatalln("[-]Read error:", err)
			}
			break
		}
		if count > 0 {
			output.Write(buf[:count])
		}
	}
	return
}
