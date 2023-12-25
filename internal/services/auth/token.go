package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenClaims struct {
	GUID      string `json:"guid"`
	SessionID string `json:"session_id"`
	jwt.StandardClaims
}

type tokenType string

const (
	accessTokenType  tokenType = "access_token"
	refreshTokenType tokenType = "refresh_token"
)

type Token = string

func (a *Service) CreateToken(tokenType tokenType, claims jwt.Claims) (Token, error) {
	var (
		iat       = time.Now()
		expiresAt int64
	)

	switch tokenType {
	case accessTokenType:
		expiresAt = iat.Add(a.config.AccessTokenExpiration).Unix()

	case refreshTokenType:
		expiresAt = iat.Add(a.config.RefreshTokenExpiration).Unix()

	default:
		return "", errors.New("unexpected token type =" + string(tokenType))
	}

	tokenClaims, ok := claims.(TokenClaims)
	if !ok {
		return "", errors.New("can't convert claims to TokenClaims")
	}

	tokenClaims.IssuedAt = iat.Unix()
	tokenClaims.ExpiresAt = expiresAt
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, tokenClaims)
	jwtToken.Header["typ"] = string(tokenType)

	t, err := jwtToken.SignedString([]byte(a.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("signed string error :%w", err)
	}

	return t, nil
}

func (a *Service) ParseAndValidate(typ tokenType, token Token) (*TokenClaims, error) {
	claims := &TokenClaims{}

	jwtToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method for jwt token")
		}

		return []byte(a.config.SecretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse with claims error %w", err)
	}

	typStr, ok := jwtToken.Header["typ"].(string)
	if !ok {
		return nil, errors.New("jwt token contains invalid header typ")
	}

	if typStr != string(typ) {
		return nil, fmt.Errorf("jwt token contains header = %s but expected = %s", jwtToken.Header["typ"], typ)
	}

	if !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
