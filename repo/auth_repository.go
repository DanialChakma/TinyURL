package repo

import (
	"context"
	"errors"

	"go.mod/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthRepository struct {
	userCol  *mongo.Collection
	tokenCol *mongo.Collection
}

func NewAuthRepository(db *mongo.Database) *AuthRepository {
	return &AuthRepository{
		userCol:  db.Collection("Users"),
		tokenCol: db.Collection("RefreshTokens"),
	}
}

// --------------------- User ---------------------
func (r *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.userCol.InsertOne(ctx, user)
	return err
}

func (r *AuthRepository) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.userCol.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) IncrementTokenVersion(ctx context.Context, userID interface{}) error {
	_, err := r.userCol.UpdateByID(ctx, userID, bson.M{"$inc": bson.M{"token_version": 1}})
	return err
}

// --------------------- RefreshToken ---------------------
func (r *AuthRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	_, err := r.tokenCol.InsertOne(ctx, token)
	return err
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, jti string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.tokenCol.FindOne(ctx, bson.M{"jti": jti}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, jti string) error {
	res, err := r.tokenCol.UpdateOne(ctx, bson.M{"jti": jti}, bson.M{"$set": bson.M{"revoked": true}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}
