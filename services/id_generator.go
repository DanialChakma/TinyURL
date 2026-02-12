package services

import (
	"errors"
	"hash/fnv"
	"log"
	"sync/atomic"
	"time"

	"go.mod/initializers"
)

const (
	nodeBits     = 10
	sequenceBits = 8
	maxNode      = -1 ^ (-1 << nodeBits)     // 1023
	maxSequence  = -1 ^ (-1 << sequenceBits) // 255
	epoch        = 1672531200000             // custom epoch in ms (Jan 1, 2023)
)

type IDGenerator struct {
	lastTs   int64
	sequence int64
	nodeID   int64
}

// -----------------------
// Constructor
// -----------------------
func NewIDGenerator(nodeID int64) *IDGenerator {
	if nodeID > maxNode || nodeID < 0 {
		panic("Node ID out of range")
	}
	return &IDGenerator{nodeID: nodeID}
}

// -----------------------
// Generate next ID (thread-safe, mostly lock-free)
// Returns uint64
// -----------------------
func (g *IDGenerator) NextID() uint64 {
	for {
		now := time.Now().UnixMilli() - epoch
		last := atomic.LoadInt64(&g.lastTs)

		if now == last {
			seq := (atomic.AddInt64(&g.sequence, 1)) & maxSequence
			if seq == 0 {
				// Wait for next millisecond
				time.Sleep(time.Millisecond)
				continue
			}
		} else {
			atomic.StoreInt64(&g.sequence, 0)
			atomic.StoreInt64(&g.lastTs, now)
		}

		// Combine parts into uint64
		id := (uint64(now) << (nodeBits + sequenceBits)) |
			(uint64(g.nodeID) << sequenceBits) |
			uint64(atomic.LoadInt64(&g.sequence))

		return id
	}
}

// -----------------------
// Feistel core (reversible)
// -----------------------
// feistelRound performs one round of a Feistel network.
// Parameters:
//   left  - left 32-bit half of the block
//   right - right 32-bit half of the block
//   key   - 32-bit round key
//
// Returns:
//   (newLeft, newRight)
//
// Operation:
//   newLeft  = right
//   newRight = left XOR F(right, key)
//
// where F(right, key) = (right + key) * 0x9e3779b9
//
// Properties:
//   - Reversible (because Feistel structure guarantees invertibility)
//   - Operates on 64-bit block split into two 32-bit halves
//   - Deterministic
//   - Not cryptographically secure (simple mixing function)
//
// Output range:
//   newLeft  ∈ [0, 2^32-1]
//   newRight ∈ [0, 2^32-1]
//
// Time complexity:
//   O(1)

// func feistelRound(left, right, key uint32) (uint32, uint32) {
// 	// 0x9e3779b9 equivalent to 2^32 / golden ratio(), for better bit-distribution it is preffered.
// 	newLeft := right
// 	newRight := left ^ ((right + key) * 0x9e3779b9)
// 	return newLeft, newRight
// }

func feistelRound(left, right, key uint32) (uint32, uint32) {
	mixed := right + key
	// mixed = bits.RotateLeft32(mixed, 7)
	mixed = (mixed << 7) | (mixed >> (32 - 7))
	// 0x9e3779b9 equivalent to 2^32 / golden ratio(phi), for better bit-distribution it is preffered.
	mixed *= 0x9e3779b9
	return right, left ^ mixed
}

// ObfuscateIDWithKey applies Feistel using a specific key
// ObfuscateIDWithKey applies multiple Feistel rounds to a 64-bit ID.
// Parameters:
//
//	id     - 64-bit unsigned integer (0 to 2^64-1)
//	key    - 32-bit base key
//	rounds - number of Feistel rounds (recommended >= 3)
//
// Returns:
//
//	uint64 obfuscated value
//
// Process:
//  1. Split id into:
//     left  = upper 32 bits
//     right = lower 32 bits
//  2. Apply Feistel rounds sequentially
//  3. Recombine halves into uint64
//
// Output range:
//
//	Always in [0, 2^64-1]
//
// IMPORTANT:
//   - Output range is identical to input range
//   - This is a permutation over 64-bit space
//   - No collisions (bijective mapping for fixed key+rounds)
//
// Time complexity:
//
//	O(rounds)
//
// Security:
//   - Provides obfuscation
//   - NOT strong cryptography
//   - Resistant to casual pattern guessing
//   - Not safe against determined cryptanalysis
func ObfuscateIDWithKey(id uint64, key uint32, rounds int) uint64 {
	left := uint32(id >> 32)
	right := uint32(id & 0xffffffff)

	for i := 0; i < rounds; i++ {
		left, right = feistelRound(left, right, key+uint32(i))
	}

	return uint64(left)<<32 | uint64(right)
}

// Feistel reverse round
func feistelReverse(left, right, key uint32) (uint32, uint32) {
	newRight := left ^ ((right + key) * 0x9e3779b9)
	newLeft := right
	return newLeft, newRight
}

// -----------------------
// Tenant key derivation
// -----------------------
func TenantKey(tenantID string) uint32 {
	if tenantID == "" || initializers.MasterFeistelSecret == "" {
		return initializers.FeistelKeys[initializers.ActiveFeistelVersion]
	}

	h := fnv.New32a()
	h.Write([]byte(tenantID))
	h.Write([]byte(initializers.MasterFeistelSecret))
	return h.Sum32()
}

// -----------------------
// Base62 encode/decode
// -----------------------
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	// Preallocate buffer for efficiency
	buf := make([]byte, 0, 11) // max int64 base62 length ~11
	for num > 0 {
		rem := num % 62
		buf = append(buf, base62Chars[rem])
		num /= 62
	}

	// Reverse in place
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func decodeBase62(str string) (int64, error) {
	var num int64
	for i := 0; i < len(str); i++ {
		var val int64
		c := str[i]
		switch {
		case '0' <= c && c <= '9':
			val = int64(c - '0')
		case 'a' <= c && c <= 'z':
			val = int64(c-'a') + 10
		case 'A' <= c && c <= 'Z':
			val = int64(c-'A') + 36
		default:
			return 0, errors.New("invalid base62 character")
		}
		num = num*62 + val
	}
	return num, nil
}

// -----------------------
// Encode/Decode IDs (feistel + base62)
// -----------------------
// func EncodeID(id int64, tenantID string) string {
// 	key := initializers.FeistelKeys[initializers.ActiveFeistelVersion]
// 	if tenantID != "" {
// 		key = TenantKey(tenantID)
// 	}

// 	obf := ObfuscateIDWithKey(id, key, initializers.FeistelRounds)
// 	return EncodeBase62(obf)
// }

func EncodeID(id uint64, tenantID string) string {
	key := initializers.FeistelKeys[initializers.ActiveFeistelVersion]
	if tenantID != "" {
		key = TenantKey(tenantID)
	}

	obf := ObfuscateIDWithKey(id, key, initializers.FeistelRounds)
	encoded := EncodeBase62(uint64(obf)) // cast to uint64

	if encoded == "" {
		log.Printf("EncodeID returned empty string: id=%d, tenantID=%s, key=%d, obf=%d\n",
			id, tenantID, key, obf)
	}

	return encoded
}

func DecodeID(encoded string, tenantID string) (int64, error) {
	num, err := decodeBase62(encoded)
	if err != nil {
		return 0, err
	}

	key := initializers.FeistelKeys[initializers.ActiveFeistelVersion]
	if tenantID != "" {
		key = TenantKey(tenantID)
	}

	left := uint32(num >> 32)
	right := uint32(num & 0xffffffff)

	for i := initializers.FeistelRounds - 1; i >= 0; i-- {
		left, right = feistelReverse(left, right, key+uint32(i))
	}

	return int64(left)<<32 | int64(right), nil
}
