package services

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var out bytes.Buffer
	_, err = io.Copy(&out, r)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func EncryptAES(plaintext []byte, secret string) ([]byte, error) {
	key := []byte(secret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Return nonce + ciphertext
	return append(nonce, ciphertext...), nil
}

func DecryptAES(cipherData []byte, secret string) ([]byte, error) {
	key := []byte(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherData) < nonceSize {
		return nil, errors.New("invalid cipher data")
	}

	nonce := cipherData[:nonceSize]
	ciphertext := cipherData[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func GenerateHMAC(data []byte, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return h.Sum(nil)
}

func VerifyHMAC(data, receivedMAC []byte, secret string) bool {
	expectedMAC := GenerateHMAC(data, secret)
	return hmac.Equal(expectedMAC, receivedMAC)
}
