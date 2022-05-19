package domain

import "time"

type Message struct {
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
