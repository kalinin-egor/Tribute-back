package entities

import (
	"time"

	"github.com/google/uuid"
)

// Subscription represents a subscription entity.
type Subscription struct {
	ID              uuid.UUID
	ChannelID       uuid.UUID
	UserID          int64
	ChannelUsername string
	Title           string
	Description     string
	ButtonText      string
	Price           float64
	CreatedDate     time.Time
}
