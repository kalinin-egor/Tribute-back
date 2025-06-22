package entities

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a payment entity.
type Payment struct {
	ID          uuid.UUID
	UserID      int64
	Description string
	CreatedDate time.Time
}
