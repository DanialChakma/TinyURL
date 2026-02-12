package initializers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	authSource := os.Getenv("DB_AUTH_SOURCE")

	if authSource == "" {
		authSource = "admin"
	}

	if user == "" || pass == "" || host == "" || dbName == "" {
		log.Fatal("MongoDB env variables missing")
	}

	userEnc := url.QueryEscape(user)
	passEnc := url.QueryEscape(pass)

	DBURL := fmt.Sprintf(
		"mongodb://%s:%s@%s/%s?authSource=%s",
		userEnc,
		passEnc,
		host,
		dbName,
		authSource,
	)

	log.Println("MongoDB URL constructed safely")
	client, err := mongo.NewClient(options.Client().ApplyURI(DBURL))
	if err != nil {
		log.Fatalf("MongoDB client creation failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}

	DB = client.Database(dbName)
	log.Println("Connected to MongoDB successfully")
}
