package models

type URL struct {
	ShortCode string `bson:"short_code" json:"short_code"`
	LongURL   string `bson:"long_url" json:"long_url"`
	TenantID  string `bson:"tenant_id" json:"tenant_id"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
}
