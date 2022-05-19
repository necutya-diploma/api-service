package domain

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	AccessToken  string
	RefreshToken string
}

type TokenInfo struct {
	UserID    string
	TokenID   string
	SessionID string
	PlanID    string
}

type Session struct {
	ID           uuid.UUID
	RefreshToken string
	Client       string
	IpAddress    string

	ExpiresAt time.Time
	CreatedAt time.Time
}
