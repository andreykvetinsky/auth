package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/andreykvetinsky/auth/config"
	"github.com/andreykvetinsky/auth/internal/handlers"
	"github.com/andreykvetinsky/auth/internal/repositories"
	"github.com/andreykvetinsky/auth/internal/services/auth"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		logger.Fatal("init logger failed", zap.Error(err))

		return
	}

	cfg := config.NewConfig()

	authdb, err := setupAuthDB(ctx, cfg)
	if err != nil {
		logger.Fatal("setting up auth db failed", zap.Error(err))

		return
	}

	authService := auth.NewService(cfg, authdb)
	server := handlers.NewServer(cfg, authService, logger)

	if err := server.Serve(); err != nil {
		logger.Fatal("http serve failed", zap.Error(err))
	}
}

func setupAuthDB(ctx context.Context, cfg *config.Config) (*repositories.AuthDB, error) {
	authdb, err := repositories.NewAuthDB(ctx, cfg.Mongo)
	if err != nil {
		return nil, fmt.Errorf("new auth db error: %w", err)
	}

	return authdb, nil
}
