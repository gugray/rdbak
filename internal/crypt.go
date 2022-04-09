package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func Encrypt(msg string, keyHex string) string {

	key, _ := hex.DecodeString(keyHex)
	plaintext := []byte(msg)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func Decrypt(encrypted string, keyHex string) string {

	key, _ := hex.DecodeString(keyHex)
	enc, _ := hex.DecodeString(encrypted)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s", plaintext)
}
