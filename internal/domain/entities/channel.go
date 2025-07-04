package entities

import "github.com/google/uuid"

// Channel represents a channel entity.
type Channel struct {
	ID              uuid.UUID
	UserID          int64
	ChannelTitle    string
	ChannelUsername string
	IsVerified      bool
}
