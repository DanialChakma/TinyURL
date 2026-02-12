package dto

type CreateTenantDTO struct {
	ID        string                 `json:"tenant_id" binding:"required"`
	Name      string                 `json:"tenant_name" binding:"required"`
	BaseURL   string                 `json:"base_url" binding:"required"`
	MetaDatas map[string]interface{} `json:"meta_datas"`
	Active    bool                   `json:"active"`
}

type UpdateTenantDTO struct {
	Name      *string                `json:"tenant_name,omitempty"`
	BaseURL   *string                `json:"base_url,omitempty"`
	MetaDatas map[string]interface{} `json:"meta_datas,omitempty"`
	Active    *bool                  `json:"active,omitempty"`
}
