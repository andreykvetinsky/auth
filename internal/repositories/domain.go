package repositories

import (
	"time"

	"github.com/google/uuid"
)

type TokenHash struct {
	UserID    string    `bson:"user_id"`
	SessionID uuid.UUID `bson:"session_id"`
	TokenHash []byte    `bson:"token_hash"`
	CreatedAt time.Time `bson:"created_at"`
}
