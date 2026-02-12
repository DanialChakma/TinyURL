package app

import (
	"log"
	"os"
	"strconv"

	"go.mod/initializers"
	"go.mod/services"
)

func Bootstrap() *Container {

	initializers.LoadEnvVariables()
	initializers.LoadConfig()
	initializers.LoadCryptoConfig()
	initializers.ConnectDB()
	initializers.RunMigrations()
	// auth.InitAuthService(initializers.DB) // pass your MongoDB db object
	// Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	cache := services.NewCache(redisAddr)

	// Snowflake
	nodeIDStr := os.Getenv("NODE_ID")
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid NODE_ID: %v", err)
	}

	idGen := services.NewIDGenerator(nodeID)

	return &Container{
		DB:    initializers.DB,
		Cache: cache,
		IDGen: idGen,
	}
}
