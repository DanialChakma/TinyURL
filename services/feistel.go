package services

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"math/big"
	"sync"
	"unicode/utf8"
)

// const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// -----------------------------
// Internal Feistel round function (stateless)
// -----------------------------
func feistelRoundStateless(half []byte, key []byte, round int) []byte {
	out := make([]byte, len(half))
	for i := 0; i < len(half); i++ {
		k := key[i%len(key)]
		out[i] = half[i] + k + byte(round)
	}
	return out
}

// -----------------------------
// Stateless obfuscation / deobfuscation
// -----------------------------

// ObfuscateBytes applies Feistel rounds on a byte slice using the given key.
func ObfuscateBytes(data []byte, key string, rounds int) []byte {
	if len(data) == 0 {
		return data
	}

	// Pad if odd length
	if len(data)%2 != 0 {
		data = append(data, 0)
	}

	left := make([]byte, len(data)/2)
	right := make([]byte, len(data)/2)
	copy(left, data[:len(data)/2])
	copy(right, data[len(data)/2:])

	keyBytes := []byte(key)

	for i := 0; i < rounds; i++ {
		fOut := feistelRoundStateless(right, keyBytes, i)

		// Parallel XOR for large slices
		wg := sync.WaitGroup{}
		for j := 0; j < len(left); j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				left[j] ^= fOut[j%len(fOut)]
			}(j)
		}
		wg.Wait()

		// Swap halves
		left, right = right, left
	}

	return append(left, right...)
}

// DeobfuscateBytes reverses the Feistel obfuscation
func DeobfuscateBytes(data []byte, key string, rounds int) []byte {
	if len(data) == 0 {
		return data
	}

	if len(data)%2 != 0 {
		data = append(data, 0)
	}

	left := make([]byte, len(data)/2)
	right := make([]byte, len(data)/2)
	copy(left, data[:len(data)/2])
	copy(right, data[len(data)/2:])

	keyBytes := []byte(key)

	for i := rounds - 1; i >= 0; i-- {
		// Swap halves
		left, right = right, left
		fOut := feistelRoundStateless(right, keyBytes, i)

		// Parallel XOR
		wg := sync.WaitGroup{}
		for j := 0; j < len(left); j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				left[j] ^= fOut[j%len(fOut)]
			}(j)
		}
		wg.Wait()
	}

	return append(left, right...)
}

// -----------------------------
// Helper: convert int64 to 8-byte slice
// -----------------------------
func Int64ToBytes(val int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(val))
	return buf
}

// Helper: convert 8-byte slice to int64
func BytesToInt64(buf []byte) int64 {
	if len(buf) != 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(buf))
}

// -----------------------------
// Base62 Encode / Decode
// -----------------------------
func Base62Encode(data []byte) string {
	num := new(big.Int).SetBytes(data)
	if num.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}

	result := ""
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = string(base62Chars[mod.Int64()]) + result
	}

	return result
}

func Base62Decode(s string) ([]byte, error) {
	num := big.NewInt(0)
	base := big.NewInt(62)

	for i := 0; i < len(s); i++ {
		c := s[i]
		var idx int64 = -1
		switch {
		case '0' <= c && c <= '9':
			idx = int64(c - '0')
		case 'a' <= c && c <= 'z':
			idx = int64(c - 'a' + 10)
		case 'A' <= c && c <= 'Z':
			idx = int64(c - 'A' + 36)
		default:
			return nil, errors.New("invalid Base62 character")
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(idx))
	}

	return num.Bytes(), nil
}

// -----------------------------
// Optional UTF-8 safe trimming
// -----------------------------
func BytesToStringSafe(data []byte) string {
	data = bytes.TrimRight(data, "\x00") // remove padding
	if utf8.Valid(data) {
		log.Printf("Decoded Byte is utf-8 valid")
		return string(data)
	}
	// fallback: replace invalid UTF-8
	log.Printf("Decoded Byte is utf-8 invalid")
	return string(bytes.Runes(data))
}

// func Compress(data []byte) ([]byte, error) {
// 	var buf bytes.Buffer
// 	writer := zlib.NewWriter(&buf)

// 	_, err := writer.Write(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	writer.Close()
// 	return buf.Bytes(), nil
// }

// func Decompress(data []byte) ([]byte, error) {
// 	buf := bytes.NewReader(data)
// 	reader, err := zlib.NewReader(buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer reader.Close()

// 	var out bytes.Buffer
// 	_, err = io.Copy(&out, reader)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return out.Bytes(), nil
// }
