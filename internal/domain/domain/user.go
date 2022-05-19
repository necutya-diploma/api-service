package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	AdminRole Role = "admin"
	BasicRole Role = "basic"
)

type Role string

type UserCredential struct {
	ExternalToken string `bson:"external_token"`
}

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Password  string

	ReceiveNotification bool
	IsConfirmed         bool

	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastVisitAt time.Time

	Role       Role
	Credential *UserCredential

	PlanID string

	TodayInternalRequest int
	TodayExternalRequest int
}

func (u *User) IsAdmin() bool {
	return u.Role == AdminRole
}

type UserCredentials struct {
	ID                uuid.UUID
	ExternalSecretKey string
	AccountType       string
}
