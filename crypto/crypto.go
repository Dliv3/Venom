package crypto

import (
	"github.com/Dliv3/Venom/global"
)

// InitEncryption generate secret key、protocal separator、protocal feature
func InitEncryption(passwd string) {
	if passwd != "" {
		global.SECRET_KEY = Md5Raw(passwd)
		global.PROTOCOL_SEPARATOR = string(Md5Raw(passwd + global.PROTOCOL_SEPARATOR)[:4])
		global.PROTOCOL_FEATURE = string(Md5Raw(passwd + global.PROTOCOL_FEATURE)[:8])
	} else {
		// 加密算法导致的缓冲区额外开销
		OVERHEAD = 0
	}
}
