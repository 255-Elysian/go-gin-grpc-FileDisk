package token

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}
