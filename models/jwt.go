package models

import "github.com/golang-jwt/jwt/v4"

type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	jwt.RegisteredClaims
}
