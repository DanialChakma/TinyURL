package repo

import (
	"context"

	"go.mod/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id string) (*models.Tenant, error)
	GetAll(ctx context.Context) ([]models.Tenant, error)
	Update(ctx context.Context, id string, update bson.M) error
	Delete(ctx context.Context, id string) error
}

type tenantRepository struct {
	collection *mongo.Collection
}

func NewTenantRepository(db *mongo.Database) TenantRepository {
	return &tenantRepository{
		collection: db.Collection("tenants"),
	}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	_, err := r.collection.InsertOne(ctx, tenant)
	return err
}

func (r *tenantRepository) GetByID(ctx context.Context, id string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.collection.FindOne(ctx, bson.M{"tenant_id": id}).Decode(&tenant)
	return &tenant, err
}

func (r *tenantRepository) GetAll(ctx context.Context) ([]models.Tenant, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tenants []models.Tenant
	for cursor.Next(ctx) {
		var t models.Tenant
		if err := cursor.Decode(&t); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, nil
}

func (r *tenantRepository) Update(ctx context.Context, id string, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"tenant_id": id}, update)
	return err
}

func (r *tenantRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"tenant_id": id})
	return err
}
