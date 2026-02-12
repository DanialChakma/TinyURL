package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mod/models"
	"go.mod/repo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   *repo.AuthRepository
	tokens *TokenService
}

func NewAuthService(repo *repo.AuthRepository, tokens *TokenService) *AuthService {
	return &AuthService{repo: repo, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, user *models.User) (*models.User, error) {
	existing, _ := s.repo.FindUserByUsername(ctx, user.Username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashed)
	user.TokenVersion = 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err := s.repo.CreateUser(ctx, user)
	user.Password = "" // do not return hashed password
	return user, err
}

func (s *AuthService) Login(ctx context.Context, username, password string) (access, refresh string, err error) {
	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil || user == nil {
		return "", "", errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", "", errors.New("invalid credentials")
	}

	access, _ = s.tokens.GenerateJWT(user.Username, user.Role, user.TokenVersion)
	jti := uuid.New().String()
	refresh, _ = s.tokens.GenerateRefreshJWT(user.Username, user.Role, jti)

	rt := &models.RefreshToken{
		UserID:    user.ID,
		JTI:       jti,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	_ = s.repo.SaveRefreshToken(ctx, rt)
	return
}

// Add Refresh method similarly
func (s *AuthService) Refresh(ctx context.Context, oldRefreshToken string) (string, string, error) {
	// 1️⃣ Validate refresh token JWT
	claims, err := s.tokens.ValidateRefreshJWT(oldRefreshToken)
	if err != nil {
		return "", "", errors.New("invalid or expired refresh token")
	}

	// 2️⃣ Get refresh token from DB using JTI
	rt, err := s.repo.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		return "", "", errors.New("failed to fetch refresh token")
	}
	if rt == nil || rt.Revoked || rt.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token revoked or expired")
	}

	// 3️⃣ Fetch user
	user, err := s.repo.FindUserByUsername(ctx, claims.Username)
	if err != nil || user == nil {
		return "", "", errors.New("user not found")
	}

	// 4️⃣ Revoke old refresh token (rotation)
	if err := s.repo.RevokeRefreshToken(ctx, claims.ID); err != nil {
		return "", "", errors.New("failed to revoke old refresh token")
	}

	// 5️⃣ Generate new tokens
	accessToken, err := s.tokens.GenerateJWT(user.Username, user.Role, user.TokenVersion)
	if err != nil {
		return "", "", errors.New("failed to generate access token")
	}

	newJTI := uuid.New().String()
	refreshToken, err := s.tokens.GenerateRefreshJWT(user.Username, user.Role, newJTI)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}

	// 6️⃣ Save new refresh token
	newRT := &models.RefreshToken{
		UserID:    user.ID,
		JTI:       newJTI,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveRefreshToken(ctx, newRT); err != nil {
		return "", "", errors.New("failed to store new refresh token")
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.FindUserByUsername(ctx, username)
}
