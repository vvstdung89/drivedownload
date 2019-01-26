package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

func EncryptString(cryptoText, keyStr string) string {
	if keyStr == "" {
		keyStr = "defaultpass"
	}

	keyBytes := sha256.Sum256([]byte(keyStr))
	return aesEncrypt(keyBytes[:], cryptoText)
}

// encrypt string to base64 crypto using AES
func aesEncrypt(key []byte, text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("AES something wrong 1")
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		fmt.Println("AES something wrong 2")
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return hex.EncodeToString(ciphertext)
}

func DecryptString(cryptoText, keyStr string) string {
	if keyStr == "" {
		keyStr = "defaultpass"
	}
	keyBytes := sha256.Sum256([]byte(keyStr))
	return aesDecrypt(keyBytes[:], cryptoText)
}

// decrypt from hexstring to decrypted string
func aesDecrypt(key []byte, cryptoText string) string {
	ciphertext, err := hex.DecodeString(cryptoText)
	if err != nil {
		fmt.Println("AES something wrong 3")
		return ""
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("AES something wrong 4")
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		fmt.Println("AES something wrong 5")
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext)
}
