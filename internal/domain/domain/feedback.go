package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	feedbackEmailTemplate = `CheckIT feedback:
From: %s \n
Email: %s \n
Text: %s \n
Created at: %s \n
Is system user: %v
`
)

type Feedback struct {
	ID uuid.UUID

	UserID string

	Username string
	Email    string

	Text      string
	Processed bool

	CreatedAt  time.Time
	ResolvedAt *time.Time
}

func (f *Feedback) GenerateEmailSubject() string {
	return fmt.Sprintf("CheckIT feedback from %s", f.Email)
}

func (f *Feedback) GenerateEmailBody() string {
	return fmt.Sprintf(
		feedbackEmailTemplate,
		f.Username,
		f.Email,
		f.Text,
		f.CreatedAt.Format(time.RFC822),
		f.UserID != "",
	)
}
