package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/andreykvetinsky/auth/config"
	"github.com/andreykvetinsky/auth/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	config *config.Config

	authRepository Repository
}

func NewService(
	config *config.Config,
	repository Repository,
) *Service {
	return &Service{
		config:         config,
		authRepository: repository,
	}
}

func (a *Service) GetTokens(ctx context.Context, userID string) (*domain.AuthResponse, error) {
	sessionID := uuid.New()

	tokenClaims := TokenClaims{
		GUID:      userID,
		SessionID: sessionID.String(),
	}

	access, err := a.CreateToken(accessTokenType, tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("create access token error %w", err)
	}

	refresh, err := a.CreateToken(refreshTokenType, tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("create refresh token error %w", err)
	}

	hash, err := makeBcryptHash(refresh)
	if err != nil {
		return nil, fmt.Errorf("generate hash refresh token error %w", err)
	}

	if err := a.authRepository.SaveHashRefreshToken(ctx, userID, sessionID, hash); err != nil {
		return nil, fmt.Errorf("save hash refresh token error %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (a *Service) RefereshTokens(ctx context.Context, accessToken, refreshToken string) (*domain.AuthResponse, error) {
	access, err := a.ParseAndValidate(accessTokenType, accessToken)
	if err != nil {
		return nil, fmt.Errorf("parse and validate access token error %w", err)
	}

	refresh, err := a.ParseAndValidate(refreshTokenType, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("parse and validate refresh token error %w", err)
	}

	if refresh.GUID != access.GUID {
		return nil, errors.New("mismatsh guid in access and refresh tokens")
	}

	if refresh.SessionID != access.SessionID {
		return nil, errors.New("mismatsh session id in access and refresh tokens")
	}

	sessionID, err := uuid.Parse(refresh.SessionID)
	if err != nil {
		return nil, fmt.Errorf("parse session_id error %w", err)
	}

	savedHash, err := a.authRepository.SearchHashRefreshToken(ctx, sessionID, refresh.GUID)
	if err != nil {
		return nil, fmt.Errorf("search hash refresh token error %w", err)
	}

	err = compareTokenWithbcrypthash(refreshToken, savedHash)
	if err != nil {
		return nil, fmt.Errorf("compare token with bcrypt hash error %w", err)
	}

	response, err := a.GetTokens(ctx, access.GUID)
	if err != nil {
		return nil, fmt.Errorf("generate tokens error %w", err)
	}

	return response, nil
}

func makeBcryptHash(token string) ([]byte, error) {
	sha256Hash := sha256.Sum256([]byte(token))
	bcryptHash, err := bcrypt.GenerateFromPassword(sha256Hash[:], bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("generate from password %w", err)
	}

	return bcryptHash, nil
}

func compareTokenWithbcrypthash(token string, hash []byte) error {
	sha256Hash := sha256.Sum256([]byte(token))

	if err := bcrypt.CompareHashAndPassword(hash, sha256Hash[:]); err != nil {
		return fmt.Errorf("compare hash and password %w", err)
	}

	return nil
}
