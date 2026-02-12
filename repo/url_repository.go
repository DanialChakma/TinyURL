package repo

import (
	"context"

	"go.mod/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type URLRepository struct {
	collection *mongo.Collection
}

func NewURLRepository(db *mongo.Database) *URLRepository {
	return &URLRepository{
		collection: db.Collection("urls"),
	}
}

func (r *URLRepository) Create(ctx context.Context, url *models.URL) error {
	_, err := r.collection.InsertOne(ctx, url)
	return err
}

func (r *URLRepository) FindByShortCode(ctx context.Context, code string) (*models.URL, error) {
	var result models.URL
	err := r.collection.
		FindOne(ctx, bson.M{"short_code": code}).
		Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
