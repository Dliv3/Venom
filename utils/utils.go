package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"runtime"
	"time"
)

/* -------------------------- encoding/decoding ------------------------------ */

// UUIDToArray32 UUID to 32byte arrary
func UUIDToArray32(hashID string) [32]byte {
	currnetHashIDSlice, _ := hex.DecodeString(hashID)
	return SliceTo32ByteArray(currnetHashIDSlice)
}

// Array32ToUUID 32byte array to UUID
func Array32ToUUID(bytes [32]byte) string {
	return fmt.Sprintf("%x", bytes)
}

/* -----------------------  Hash/Random number --------------------------- */

// NewUUID New Universally Unique Identifier
// TODO use https://github.com/satori/go.uuid to generate UUID
func NewUUID() string {
	return Sha256(time.Now().Format(time.ANSIC) + GetRandomString(5))
}

// Sha256 Hash
func Sha256(text string) string {
	ctx := sha256.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// Md5 Hash
func Md5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// GetRandomString generate a random string
func GetRandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

/* -----------------------  Type conversion --------------------------- */

func Uint32ToBytes(i uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func BytesToUint32(buf []byte) uint32 {
	return uint32(binary.BigEndian.Uint32(buf))
}

func Uint16ToBytes(i uint16) []byte {
	var buf = make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(i))
	return buf
}

func BytesToUint16(buf []byte) uint16 {
	return uint16(binary.BigEndian.Uint16(buf))
}

func Uint64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToUint64(buf []byte) uint64 {
	return uint64(binary.BigEndian.Uint64(buf))
}

func SliceTo32ByteArray(buf []byte) [32]byte {
	var result [32]byte
	for i := 0; i < 32; i++ {
		result[i] = buf[i]
	}
	return result
}

func IpToUint32(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func Uint32ToIp(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

/* -----------------------  system info --------------------------- */
func GetSystemType() (osType uint32) {
	var os = runtime.GOOS
	switch os {
	case "darwin":
		osType = 0x00
	case "windows":
		osType = 0x01
	case "linux":
		osType = 0x02
	default:
		// unknown
		osType = 0xff
	}
	return
}

/* ------------------------ file operation ---------------------------- */

// FileExists Check if the file exists
func FileExists(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// GetFileSize Get the size of a single file
func GetFileSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size() //获取size
	return fileSize
}

/* -------------- Calculate structure size ----------------- */
func PacketSize(packet interface{}) (uint64, error) {

	var size uint64

	size = 0

	t := reflect.TypeOf(packet)
	v := reflect.ValueOf(packet)

	if k := t.Kind(); k != reflect.Struct {
		return 0, errors.New("Param is Not Struct")
	}

	count := t.NumField()
	for i := 0; i < count; i++ {
		val := v.Field(i).Interface()

		// type switch
		switch value := val.(type) {
		case uint16:
			size += 2
		case uint32:
			size += 4
		case uint64:
			size += 8
		case string:
			size += uint64(len(value))
		case []byte:
			size += uint64(len(value))
		case [2]byte:
			size += uint64(len(value))
		case [4]byte:
			size += uint64(len(value))
		case [32]byte:
			size += uint64(len(value))
		default:
			log.Fatalln("[-]PacketSize error: type unsupport")
			return 0, errors.New("Type unsupport")
		}
	}
	return size, nil
}

/* ------------------- slice deduplication ---------------------*/
func RemoveDuplicateElement(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	temp := map[string]struct{}{}
	for _, item := range addrs {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

/* --------------------- handling windows \r ------------------------------*/
func HandleWindowsCR() {
	if runtime.GOOS == "windows" {
		var noUse string
		fmt.Scanf("%s", &noUse)
		// fmt.Println("windows :", []byte(noUse))
	}
}
