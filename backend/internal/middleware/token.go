package middleware

import "RealTimeChat/backend/internal/helpers"

type TokenProvider interface {
	Validate(token string) (*helpers.Claims, error)
}

type JWTTokenProvider struct{}

func NewJWTTokenProvider() TokenProvider {
	return &JWTTokenProvider{}
}

func (p *JWTTokenProvider) Validate(token string) (*helpers.Claims, error) {
	return helpers.ValidateToken(token)
}
