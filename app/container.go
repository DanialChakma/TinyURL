package app

import (
	"go.mod/services"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	DB    *mongo.Database
	Cache *services.Cache
	IDGen *services.IDGenerator
}
