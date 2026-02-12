package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/dto"
	"go.mod/services"
)

type TenantController struct {
	service *services.TenantService
}

func NewTenantController(service *services.TenantService) *TenantController {
	return &TenantController{service: service}
}

// =====================
// CREATE
// =====================
// CreateTenant godoc
// @Summary Create a new tenant
// @Description Create tenant (Admin only)
// @Tags Tenants
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param tenant body dto.CreateTenantDTO true "Tenant data"
// @Success 201 {object} models.Tenant
// @Failure 400 {object} map[string]string
// @Router /tenants [post]
func (tc *TenantController) CreateTenant(c *gin.Context) {

	var req dto.CreateTenantDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := tc.service.CreateTenant(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// =====================
// GET BY ID
// =====================

// GetTenant godoc
// @Summary Get tenant by ID
// @Description Retrieve a tenant using tenant ID
// @Tags Tenants
// @Security BearerAuth
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} models.Tenant
// @Failure 404 {object} map[string]string
// @Router /tenants/{id} [get]
func (tc *TenantController) GetTenant(c *gin.Context) {

	id := c.Param("id")

	tenant, err := tc.service.GetTenantByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// =====================
// GET ALL
// =====================

// GetAllTenants godoc
// @Summary Get all tenants
// @Description Retrieve all tenants
// @Tags Tenants
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Tenant
// @Failure 500 {object} map[string]string
// @Router /tenants [get]
func (tc *TenantController) GetAllTenants(c *gin.Context) {

	tenants, err := tc.service.GetAllTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenants)
}

// =====================
// UPDATE
// =====================

// UpdateTenant godoc
// @Summary Update tenant
// @Description Update tenant by ID
// @Tags Tenants
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param tenant body dto.UpdateTenantDTO true "Updated tenant data"
// @Success 200 {object} models.Tenant
// @Failure 400 {object} map[string]string
// @Router /tenants/{id} [put]
func (tc *TenantController) UpdateTenant(c *gin.Context) {

	id := c.Param("id")

	var req dto.UpdateTenantDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := tc.service.UpdateTenant(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// =====================
// DELETE
// =====================

// DeleteTenant godoc
// @Summary Delete tenant
// @Description Delete tenant by ID
// @Tags Tenants
// @Security BearerAuth
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tenants/{id} [delete]
func (tc *TenantController) DeleteTenant(c *gin.Context) {

	id := c.Param("id")

	err := tc.service.DeleteTenant(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tenant deleted successfully"})
}
