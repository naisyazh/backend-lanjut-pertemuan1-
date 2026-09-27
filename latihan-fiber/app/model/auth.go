package model

import "time"

// RegisterRequest adalah request body untuk POST /auth/register
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
	// TIDAK ADA field Role - role ditentukan server untuk keamanan
}

// LoginRequest adalah request body untuk POST /auth/login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest adalah request body untuk POST /auth/refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenPair adalah response untuk login dan refresh
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"` // Always "Bearer"
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// RefreshToken adalah entitas refresh token di database
type RefreshToken struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"` // Never expose hash
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// AuthUser adalah identitas user yang dibawa dalam access token
type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}