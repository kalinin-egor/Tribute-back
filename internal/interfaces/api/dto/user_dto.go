package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	Username  *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user data in HTTP responses
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginResponse represents the response body for user login
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// ToUserResponse converts domain entity to DTO
func ToUserResponse(user interface{}) UserResponse {
	// This would be implemented based on your domain entity
	// For now, it's a placeholder
	return UserResponse{}
}

// Dashboard
type DashboardRequest struct {
	AccessToken string `json:"access_token"`
}

type DashboardResponse struct {
	Earn              float64      `json:"earn"`
	ChannelsAndGroups []ChannelDTO `json:"channels-and-groups"`
	IsVerified        bool         `json:"is-verified"`
	Subscriptions     []SubDTO     `json:"subscriptions"`
	IsSubPublished    bool         `json:"is-sub-published"`
	PaymentsHistory   []PaymentDTO `json:"payments-history"`
	IsOnboarded       bool         `json:"is-onboarded"`
}

// AddBot
type AddBotRequest struct {
	AccessToken string `json:"access_token"`
	BotUsername string `json:"bot-username"`
}

// UploadVerifiedPassport
type UploadVerifiedPassportRequest struct {
	AccessToken  string `json:"access_token"`
	UserPhoto    string `json:"user-photo"`    // Assuming base64 encoded string
	UserPassport string `json:"user-passport"` // Assuming base64 encoded string
}

// CheckVerifiedPassport
type CheckVerifiedPassportRequest struct {
	UserID int64 `json:"user_id"`
}

type CheckVerifiedPassportResponse struct {
	UserID     int64 `json:"user_id"`
	IsVerified bool  `json:"is_verified"`
}

// SetUpPayouts
type SetUpPayoutsRequest struct {
	AccessToken string `json:"access_token"`
	CardNumber  string `json:"card-number"`
	CardDate    string `json:"card-date"`
	CardCVV     string `json:"card-cvv"`
}

// PublishSubscription
type PublishSubscriptionRequest struct {
	AccessToken string  `json:"access_token"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ButtonText  string  `json:"button-text"`
	Price       float64 `json:"price"`
}

// CreateSubscribe
type CreateSubscribeRequest struct {
	UserID int64   `json:"user_id"`
	Price  float64 `json:"price"`
}

// --- Reusable DTOs ---
type ChannelDTO struct {
	ID              uuid.UUID `json:"id"`
	ChannelUsername string    `json:"channel_username"`
}

type SubDTO struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
}

type PaymentDTO struct {
	Description string `json:"description"`
	CreatedDate string `json:"created-date"`
}
