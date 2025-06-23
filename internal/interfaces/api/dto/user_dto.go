package dto

import (
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
	CardNumber        string       `json:"card_number"`
}

// AddBot
type AddBotRequest struct {
	UserID          int64  `json:"user_id" binding:"required"`
	ChannelTitle    string `json:"channel_title" binding:"required"`
	ChannelUsername string `json:"channel_username" binding:"required"`
}

// CheckChannelRequest represents the request for checking channel ownership
type CheckChannelRequest struct {
	ChannelID uuid.UUID `json:"channel_id" binding:"required"`
}

// CheckChannelResponse represents the response for channel ownership check
type CheckChannelResponse struct {
	IsOwner bool `json:"is_owner"`
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
	CardNumber string `json:"card-number" binding:"required"`
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
	ChannelTitle    string    `json:"channel_title"`
	ChannelUsername string    `json:"channel_username"`
	IsVerified      bool      `json:"is_verified"`
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

// --- Generic Response DTOs ---

// MessageResponse is a generic response for a successful operation.
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is a generic response for an error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse is a generic response for a status message.
type StatusResponse struct {
	Status string `json:"status"`
}

// --- Specific Response DTOs ---

// UserResponse represents a user's data in a response.
type UserResponse struct {
	ID             int64   `json:"id"`
	Earned         float64 `json:"earned"`
	IsVerified     bool    `json:"is_verified"`
	IsSubPublished bool    `json:"is_sub_published"`
	IsOnboarded    bool    `json:"is_onboarded"`
	CardNumber     string  `json:"card_number"`
}

// OnboardResponse is the response for a successful onboarding.
type OnboardResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
}

// AddBotResponse is the response for successfully adding a bot.
type AddBotResponse struct {
	Message string     `json:"message"`
	Channel ChannelDTO `json:"channel"`
}

// PublishSubscriptionResponse is the response for a successful subscription publication.
type PublishSubscriptionResponse struct {
	Message      string `json:"message"`
	Subscription SubDTO `json:"subscription"`
}

// CreateUserResponse is the response for creating a user
type CreateUserResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
	Created bool         `json:"created"`
}
