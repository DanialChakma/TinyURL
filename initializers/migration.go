package initializers

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RunMigrations() {
	collection := DB.Collection("urls")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			// {Key: "tenant_id", Value: 1},
			{Key: "short_code", Value: 1}, // 1 means ascending, -1 means descending
		},
		Options: options.Index().
			SetUnique(true).
			SetBackground(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	collection = DB.Collection("tenants")

	_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "tenant_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	log.Println("MongoDB indexes ensured successfully")
}
