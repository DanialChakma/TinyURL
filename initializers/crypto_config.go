package initializers

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	FeistelKeys          []uint32
	ActiveFeistelVersion int
	FeistelRounds        int
	MasterFeistelSecret  string
	AESSecret            string
	HMACSecret           string
)

func LoadCryptoConfig() {
	keys := strings.Split(os.Getenv("FEISTEL_KEYS"), ",")
	if len(keys) == 0 {
		log.Fatal("FEISTEL_KEYS not set")
	}

	for _, k := range keys {
		v, err := strconv.ParseUint(strings.TrimSpace(k), 10, 32)
		if err != nil {
			log.Fatal("Invalid FEISTEL_KEYS value")
		}
		FeistelKeys = append(FeistelKeys, uint32(v))
	}

	v, err := strconv.Atoi(os.Getenv("ACTIVE_FEISTEL_VERSION"))
	if err != nil || v < 0 || v >= len(FeistelKeys) {
		log.Fatal("Invalid ACTIVE_FEISTEL_VERSION")
	}
	ActiveFeistelVersion = v

	// FEISTEL_ROUNDS
	// ---- FEISTEL_ROUNDS ----
	roundsEnv := os.Getenv("FEISTEL_ROUNDS")
	if roundsEnv == "" {
		FeistelRounds = 3 // safe default
	} else {
		r, err := strconv.Atoi(roundsEnv)
		if err != nil || r < 1 || r > 10 {
			log.Fatal("Invalid FEISTEL_ROUNDS (must be between 1 and 10)")
		}
		FeistelRounds = r
	}

	MasterFeistelSecret = os.Getenv("MASTER_FEISTEL_SECRET")
	if MasterFeistelSecret == "" {
		log.Println("Warning: MASTER_FEISTEL_SECRET not set. Multi-tenant obfuscation disabled.")
	}

	AESSecret = os.Getenv("AES_SECRET")
	HMACSecret = os.Getenv("HMAC_SECRET")

	if len(AESSecret) != 32 {
		log.Fatal("AES_SECRET must be exactly 32 bytes for AES-256")
	}

	if HMACSecret == "" {
		log.Fatal("HMAC_SECRET not set")
	}

}
