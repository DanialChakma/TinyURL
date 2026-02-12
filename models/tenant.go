package models

type Tenant struct {
	ID        string                 `bson:"tenant_id" json:"tenant_id"`     // unique tenant identifier (slug)
	Name      string                 `bson:"tenant_name" json:"tenant_name"` // display name
	BaseURL   string                 `bson:"base_url" json:"base_url"`       // used for redirect prefix
	MetaDatas map[string]interface{} `bson:"meta_datas,omitempty" json:"meta_datas,omitempty"`
	Active    bool                   `bson:"active" json:"active"` // enable/disable tenant
	CreatedAt int64                  `bson:"created_at" json:"created_at"`
	UpdatedAt int64                  `bson:"updated_at" json:"updated_at"`
}
