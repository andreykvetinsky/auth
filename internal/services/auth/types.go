package auth

import (
	"context"

	"github.com/google/uuid"
)

type (
	Repository interface {
		SaveHashRefreshToken(ctx context.Context, userID string, sessionID uuid.UUID, hash []byte) error
		SearchHashRefreshToken(ctx context.Context, sessionID uuid.UUID, userID string) ([]byte, error)
	}
)
