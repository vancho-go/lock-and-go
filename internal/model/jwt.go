package model

import "github.com/golang-jwt/jwt/v5"

// JWTClaims модель для JWT.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}
