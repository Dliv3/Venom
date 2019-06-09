package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

var OVERHEAD = aes.BlockSize

// reference: https://golang.org/src/crypto/cipher/example_test.go
// Special thanks to 00theway for his help. (https://github.com/00theway)

// Encrypt AES CTR
func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

// Decrypt AES CTR
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	// if len(ciphertext) < aes.BlockSize {
	// 	return nil, errors.New("ciphertext too short")
	// }

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// CTR mode is the same for both encryption and decryption, so we can
	// also decrypt that ciphertext with NewCTR
	plaintext := make([]byte, len(ciphertext[aes.BlockSize:]))
	iv := ciphertext[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])
	return plaintext, nil
}
