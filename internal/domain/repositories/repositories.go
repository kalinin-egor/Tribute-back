package repositories

import (
	"tribute-back/internal/domain/entities"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	FindByID(id int64) (*entities.User, error)
	Update(user *entities.User) error
	Create(user *entities.User) error
	// Add other necessary methods
}

// ChannelRepository defines the interface for channel data operations
type ChannelRepository interface {
	FindByUserID(userID int64) ([]*entities.Channel, error)
	FindByID(id uuid.UUID) (*entities.Channel, error)
	Create(channel *entities.Channel) error
	Update(channel *entities.Channel) error
	Delete(id uuid.UUID) error
	// Add other necessary methods
}

// SubscriptionRepository defines the interface for subscription data operations
type SubscriptionRepository interface {
	FindByID(id uuid.UUID) (*entities.Subscription, error)
	FindByUserID(userID int64) ([]*entities.Subscription, error)
	FindByChannelID(channelID uuid.UUID) (*entities.Subscription, error)
	Create(subscription *entities.Subscription) error
	Update(subscription *entities.Subscription) error
	// Add other necessary methods
}

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	FindByUserID(userID int64) ([]*entities.Payment, error)
	Create(payment *entities.Payment) error
	// Add other necessary methods
}
