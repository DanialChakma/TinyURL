package services

import (
	"context"
	"errors"
	"time"

	"go.mod/dto"
	"go.mod/models"
	"go.mod/repo"

	"go.mongodb.org/mongo-driver/bson"
)

type TenantService struct {
	repo repo.TenantRepository
}

func NewTenantService(repo repo.TenantRepository) *TenantService {
	return &TenantService{
		repo: repo,
	}
}

// =====================
// CREATE
// =====================
func (s *TenantService) CreateTenant(
	ctx context.Context,
	req dto.CreateTenantDTO,
) (*models.Tenant, error) {

	// Check if tenant already exists
	existing, err := s.repo.GetByID(ctx, req.ID)
	if err == nil && existing != nil {
		return nil, errors.New("tenant already exists")
	}

	now := time.Now().Unix()

	tenant := &models.Tenant{
		ID:        req.ID,
		Name:      req.Name,
		BaseURL:   req.BaseURL,
		MetaDatas: req.MetaDatas,
		Active:    req.Active,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

// =====================
// GET BY ID
// =====================
func (s *TenantService) GetTenantByID(
	ctx context.Context,
	id string,
) (*models.Tenant, error) {

	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	return tenant, nil
}

// =====================
// GET ALL
// =====================
func (s *TenantService) GetAllTenants(
	ctx context.Context,
) ([]models.Tenant, error) {

	return s.repo.GetAll(ctx)
}

// =====================
// UPDATE
// =====================
func (s *TenantService) UpdateTenant(
	ctx context.Context,
	id string,
	req dto.UpdateTenantDTO,
) (*models.Tenant, error) {

	// Check if tenant exists first
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	setFields := bson.M{}

	if req.Name != nil {
		setFields["tenant_name"] = *req.Name
	}
	if req.BaseURL != nil {
		setFields["base_url"] = *req.BaseURL
	}
	if req.MetaDatas != nil {
		setFields["meta_datas"] = req.MetaDatas
	}
	if req.Active != nil {
		setFields["active"] = *req.Active
	}

	setFields["updated_at"] = time.Now().Unix()

	update := bson.M{
		"$set": setFields,
	}

	if err := s.repo.Update(ctx, id, update); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

// =====================
// DELETE
// =====================
func (s *TenantService) DeleteTenant(
	ctx context.Context,
	id string,
) error {

	// Check existence first
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("tenant not found")
	}

	return s.repo.Delete(ctx, id)
}
