package initializers

import (
	"log"
	"os"
)

var (
	AppBaseURL    string
	JwtKey        []byte
	JwtRefreshKey []byte
)

func LoadConfig() {
	AppBaseURL = os.Getenv("APP_BASE_URL")
	if AppBaseURL == "" {
		log.Fatal("APP_BASE_URL not set in .env")
	}

	// JWT secrets
	JwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	JwtRefreshKey = []byte(os.Getenv("JWT_REFRESH_KEY"))
}
