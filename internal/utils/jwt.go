package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zellis-rameesn/go-ecommerce/internal/config"
)

// Claims contain data for the user
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `jsonn:"role"`
	jwt.RegisteredClaims
}

// Generate access and refresh tokens
func GenerateTokens(cfg *config.JWTConfig, userID uint, email, role string) (accessToken, refreshToken string, err error) {
	// accessToken
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	at := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims)
	accessTokenString, err := at.SignedString(cfg.Secret)
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := rt.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// Check if JWT token is valid
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
