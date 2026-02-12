package services

import (
	"errors"

	"go.mod/initializers"
)

/*
SecureURLService handles:
- Compression
- AES encryption
- HMAC signing
- Base62 encoding
- Verification
- Decryption
- Decompression
*/

type SecureURLService struct{}

func NewSecureURLService() *SecureURLService {
	return &SecureURLService{}
}

func (s *SecureURLService) Create(longURL string) (string, error) {

	// 1️⃣ Compress
	compressed, err := Compress([]byte(longURL))
	if err != nil {
		return "", errors.New("compression failed")
	}

	// 2️⃣ Encrypt
	encrypted, err := EncryptAES(compressed, initializers.AESSecret)
	if err != nil {
		return "", errors.New("encryption failed")
	}

	// 3️⃣ HMAC Sign
	signature := GenerateHMAC(encrypted, initializers.HMACSecret)

	// 4️⃣ Combine encrypted + signature
	finalPayload := append(encrypted, signature...)

	// 5️⃣ Base62 Encode
	shortCode := Base62Encode(finalPayload)

	return shortCode, nil
}

func (s *SecureURLService) Resolve(shortCode string) (string, error) {

	// 1️⃣ Base62 Decode
	data, err := Base62Decode(shortCode)
	if err != nil {
		return "", errors.New("invalid URL")
	}

	if len(data) < 32 {
		return "", errors.New("invalid token")
	}

	// 2️⃣ Split encrypted + signature
	signature := data[len(data)-32:]
	encrypted := data[:len(data)-32]

	// 3️⃣ Verify HMAC
	if !VerifyHMAC(encrypted, signature, initializers.HMACSecret) {
		return "", errors.New("tampered URL")
	}

	// 4️⃣ Decrypt
	decrypted, err := DecryptAES(encrypted, initializers.AESSecret)
	if err != nil {
		return "", errors.New("decryption failed")
	}

	// 5️⃣ Decompress
	original, err := Decompress(decrypted)
	if err != nil {
		return "", errors.New("corrupted data")
	}

	return string(original), nil
}
