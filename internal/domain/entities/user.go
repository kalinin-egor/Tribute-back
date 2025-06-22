package entities

import (
	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID             int64
	Earned         float64
	IsVerified     bool
	Subscriptions  []uuid.UUID // Assuming this holds IDs of subscriptions
	IsSubPublished bool
	IsOnboarded    bool
}
