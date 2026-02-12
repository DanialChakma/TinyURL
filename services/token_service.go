package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewTokenService(accessSecret, refreshSecret []byte) *TokenService {
	return &TokenService{accessSecret: accessSecret, refreshSecret: refreshSecret}
}

type Claims struct {
	Username     string `json:"username"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version,omitempty"`
	jwt.RegisteredClaims
}

func (t *TokenService) GenerateJWT(username, role string, tokenVersion int) (string, error) {
	exp := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Username:     username,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.accessSecret)
}

func (t *TokenService) GenerateRefreshJWT(username, role, jti string) (string, error) {
	exp := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.refreshSecret)
}

func (t *TokenService) ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return t.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

func (t *TokenService) ValidateRefreshJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return t.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
