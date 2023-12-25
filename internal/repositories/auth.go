package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andreykvetinsky/auth/config"
	"github.com/andreykvetinsky/auth/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	authDatabaseName     = "auth"
	tokensCollectionName = "tokens"
)

type AuthDB struct {
	db *mongo.Database
}

func NewAuthDB(
	ctx context.Context,
	config *config.Mongo,
) (*AuthDB, error) {
	opts := options.Client().ApplyURI(config.URI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect error: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("connect error: %w", err)
	}

	return &AuthDB{
		db: client.Database(authDatabaseName),
	}, nil
}

func (a *AuthDB) collection() *mongo.Collection {
	return a.db.Collection(tokensCollectionName)
}

func (a *AuthDB) SaveHashRefreshToken(
	ctx context.Context,
	userID string,
	sessionID uuid.UUID,
	hash []byte,
) error {
	newHash := TokenHash{
		UserID:    userID,
		TokenHash: hash,
		SessionID: sessionID,
		CreatedAt: time.Now(),
	}

	_, err := a.collection().InsertOne(ctx, newHash)
	if err != nil {
		return fmt.Errorf("insert error %w", err)
	}

	return nil
}

func (a *AuthDB) SearchHashRefreshToken(
	ctx context.Context,
	sessionID uuid.UUID,
	userID string,
) (
	[]byte,
	error,
) {
	var tokenHash TokenHash

	err := a.collection().FindOne(ctx, bson.M{"session_id": sessionID, "user_id": userID}).Decode(&tokenHash)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("find one error: %w", err)
	}

	return tokenHash.TokenHash, nil
}
