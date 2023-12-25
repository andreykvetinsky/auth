package handlers

import (
	"context"

	"github.com/andreykvetinsky/auth/internal/domain"
)

type (
	authService interface {
		GetTokens(ctx context.Context, userID string) (*domain.AuthResponse, error)
		RefereshTokens(ctx context.Context, accessToken, refreshToken string) (*domain.AuthResponse, error)
	}
)
