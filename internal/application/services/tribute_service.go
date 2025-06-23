package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"tribute-back/internal/domain/entities"
	"tribute-back/internal/domain/repositories"
	"tribute-back/internal/infrastructure/database/postgres"
	"tribute-back/internal/infrastructure/payouts"
	"tribute-back/internal/infrastructure/telegram"

	"github.com/google/uuid"
)

type TributeService struct {
	users         repositories.UserRepository
	channels      repositories.ChannelRepository
	subs          repositories.SubscriptionRepository
	payments      repositories.PaymentRepository
	telegramBot   *telegram.BotService
	payoutGateway payouts.Gateway
}

func NewTributeService(
	users repositories.UserRepository,
	channels repositories.ChannelRepository,
	subs repositories.SubscriptionRepository,
	payments repositories.PaymentRepository,
	telegramBot *telegram.BotService,
	payoutGateway payouts.Gateway,
) *TributeService {
	return &TributeService{
		users:         users,
		channels:      channels,
		subs:          subs,
		payments:      payments,
		telegramBot:   telegramBot,
		payoutGateway: payoutGateway,
	}
}

type DashboardData struct {
	User          *entities.User
	Channels      []*entities.Channel
	Subscriptions []*entities.Subscription
	Payments      []*entities.Payment
}

func (s *TributeService) GetDashboardData(userID int64) (*DashboardData, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	channels, err := s.channels.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	subscriptions, err := s.subs.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	payments, err := s.payments.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		User:          user,
		Channels:      channels,
		Subscriptions: subscriptions,
		Payments:      payments,
	}, nil
}

// SendTelegramMessage sends a message to a user via Telegram bot
func (s *TributeService) SendTelegramMessage(userID int64, message string) error {
	return s.telegramBot.SendMessage(userID, message)
}

func (s *TributeService) AddBot(userID int64, channelTitle, channelUsername string) (*entities.Channel, error) {
	// Check if the channel already exists for this user to prevent duplicates
	existingChannels, err := s.channels.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, ch := range existingChannels {
		if ch.ChannelUsername == channelUsername {
			return nil, errors.New("this channel is already added to your account")
		}
	}

	channel := &entities.Channel{
		UserID:          userID,
		ChannelTitle:    channelTitle,
		ChannelUsername: channelUsername,
		IsVerified:      false,
	}

	err = s.channels.Create(channel)
	if err != nil {
		return nil, err
	}

	// Send Telegram message after successful save
	message := fmt.Sprintf("Just a moment, we are checking bot permissions in %s", channelUsername)
	fmt.Printf("Attempting to send message to user %d: %s\n", userID, message)

	if err := s.telegramBot.SendMessage(userID, message); err != nil {
		fmt.Printf("Failed to send message to user %d: %v\n", userID, err)
	} else {
		fmt.Printf("Successfully sent message to user %d\n", userID)
	}

	return channel, nil
}

func (s *TributeService) GetChannelList(userID int64) ([]*entities.Channel, error) {
	return s.channels.FindByUserID(userID)
}

func (s *TributeService) CheckChannel(userID int64, channelID uuid.UUID) (bool, error) {
	// Get channel by ID
	channel, err := s.channels.FindByID(channelID)
	if err != nil {
		return false, err
	}
	if channel == nil {
		return false, errors.New("channel not found")
	}

	// Check if user owns this channel
	if channel.UserID != userID {
		return false, errors.New("channel does not belong to this user")
	}

	// Check if user is owner/admin of the channel via Telegram API
	chatMember, err := s.telegramBot.CheckChannelMembership(channel.ChannelUsername, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check channel membership: %w", err)
	}

	// Check if user is owner or administrator
	if chatMember.Status == "creator" || chatMember.Status == "administrator" {
		// User is owner/admin, update verification status
		channel.IsVerified = true
		if err := s.channels.Update(channel); err != nil {
			return false, fmt.Errorf("failed to update channel verification: %w", err)
		}

		// Send success message to user
		successMessage := fmt.Sprintf("Good! You added bot to channel: %s (@%s)", channel.ChannelTitle, channel.ChannelUsername)
		fmt.Printf("Attempting to send success message to user %d: %s\n", userID, successMessage)

		if err := s.telegramBot.SendMessage(userID, successMessage); err != nil {
			fmt.Printf("Failed to send success message to user %d: %v\n", userID, err)
		} else {
			fmt.Printf("Successfully sent success message to user %d\n", userID)
		}

		return true, nil
	} else {
		// User is not owner/admin, delete the channel
		if err := s.channels.Delete(channelID); err != nil {
			return false, fmt.Errorf("failed to delete channel: %w", err)
		}
		return false, nil
	}
}

func (s *TributeService) RequestVerification(userID int64, userPhotoB64, userPassportB64 string) error {
	photoReader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(userPhotoB64))
	passportReader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(userPassportB64))

	return s.telegramBot.SendVerificationRequest(userID, photoReader, passportReader)
}

func (s *TributeService) HandleVerificationCallback(chatID int64, messageID int, callbackData string) error {
	parts := strings.Split(callbackData, "_")
	if len(parts) != 3 || parts[0] != "verify" {
		return fmt.Errorf("invalid callback data format: %s", callbackData)
	}

	action := parts[1]
	userID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user id in callback data: %w", err)
	}

	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user for verification not found")
	}

	if action == "approve" {
		user.IsVerified = true
		if err := s.users.Update(user); err != nil {
			return err
		}
		// Verification successful, delete the message from admin chat
		return s.telegramBot.DeleteMessage(chatID, messageID)
	} else if action == "reject" {
		// Verification rejected, notify the user and delete message from admin chat
		if err := s.telegramBot.SendMessage(userID, "Ваша верификация была отклонена."); err != nil {
			// Log error but don't block deletion
		}
		return s.telegramBot.DeleteMessage(chatID, messageID)
	}

	return fmt.Errorf("unknown action in callback data: %s", action)
}

func (s *TributeService) SetUpPayouts(userID int64, cardNumber string) error {
	// Here you could add any business logic before contacting the payment gateway.
	// For example, check if the user is verified.
	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsVerified {
		return errors.New("user must be verified to set up payouts")
	}

	// Save card number to database
	user.CardNumber = cardNumber
	if err := s.users.Update(user); err != nil {
		return fmt.Errorf("failed to save card number to database: %w", err)
	}

	// Note: We only save the card number to our database
	// Payment gateway integration would be implemented here if needed
	return nil
}

func (s *TributeService) PublishSubscription(userID int64, title, description, buttonText string, price float64) (*entities.Subscription, error) {
	// Assumption: We use the user's first channel.
	channels, err := s.channels.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, errors.New("user has no channels to publish a subscription for")
	}
	channel := channels[0] // Use the first channel

	// Check if a subscription for this channel already exists
	subscription, err := s.subs.FindByChannelID(channel.ID)
	if err != nil {
		return nil, err
	}

	if subscription != nil {
		// Update existing subscription
		subscription.Title = title
		subscription.Description = description
		subscription.ButtonText = buttonText
		subscription.Price = price
		if err := s.subs.Update(subscription); err != nil {
			return nil, err
		}
	} else {
		// Create new subscription
		subscription = &entities.Subscription{
			ChannelID:       channel.ID,
			UserID:          userID,
			ChannelUsername: channel.ChannelUsername,
			Title:           title,
			Description:     description,
			ButtonText:      buttonText,
			Price:           price,
			CreatedDate:     time.Now(),
		}
		if err := s.subs.Create(subscription); err != nil {
			return nil, err
		}
	}

	return subscription, nil
}

// OnboardUser creates a user if they don't exist and returns dashboard data
func (s *TributeService) OnboardUser(userID int64) (*entities.User, bool, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, false, err
	}

	if user != nil {
		return user, false, nil // User already exists
	}

	// Create new user
	user = &entities.User{
		ID:          userID,
		IsVerified:  false,
		IsOnboarded: true,
	}

	if err := s.users.Create(user); err != nil {
		return nil, false, err
	}

	return user, true, nil
}

// CreateUser creates a new user
func (s *TributeService) CreateUser(userID int64) (*entities.User, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil // User already exists
	}

	// Create new user
	user = &entities.User{
		ID:          userID,
		IsVerified:  false,
		IsOnboarded: true,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateSubscription creates a subscription for a user
func (s *TributeService) CreateSubscription(subscriberID int64, creatorID int64, price float64) error {
	// Get creator's subscription
	creatorChannels, err := s.channels.FindByUserID(creatorID)
	if err != nil {
		return err
	}
	if len(creatorChannels) == 0 {
		return errors.New("creator has no channels")
	}

	creatorSubscription, err := s.subs.FindByChannelID(creatorChannels[0].ID)
	if err != nil {
		return err
	}
	if creatorSubscription == nil {
		return errors.New("creator has no subscription tier")
	}

	// Create payment record
	payment := &entities.Payment{
		ID:          uuid.New(),
		UserID:      subscriberID,
		Description: fmt.Sprintf("Subscription to user %d", creatorID),
		CreatedDate: time.Now(),
	}

	if err := s.payments.Create(payment); err != nil {
		return err
	}

	return nil
}

// ResetDatabase resets all data in the database (for development/testing)
func (s *TributeService) ResetDatabase() error {
	// Get database connection from user repository
	db := s.users.(*postgres.PgUserRepository).GetDB()

	// Drop all tables in correct order (due to foreign key constraints)
	queries := []string{
		"DROP TABLE IF EXISTS payments CASCADE",
		"DROP TABLE IF EXISTS subscriptions CASCADE",
		"DROP TABLE IF EXISTS channels CASCADE",
		"DROP TABLE IF EXISTS users CASCADE",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", query, err)
		}
	}

	// Recreate tables by running the migration
	migrationQueries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id BIGINT PRIMARY KEY,
			earned NUMERIC(10, 2) DEFAULT 0.00,
			is_verified BOOLEAN DEFAULT FALSE,
			is_sub_published BOOLEAN DEFAULT FALSE,
			is_onboarded BOOLEAN DEFAULT FALSE,
			card_number VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS channels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			channel_title VARCHAR(255) NOT NULL,
			channel_username VARCHAR(255) UNIQUE NOT NULL,
			is_verified BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			channel_username VARCHAR(255) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			button_text VARCHAR(255),
			price NUMERIC(10, 2) NOT NULL,
			created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			description TEXT,
			created_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_user_id ON channels(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_channel_id ON subscriptions(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id)`,
	}

	for _, query := range migrationQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to recreate table: %w", err)
		}
	}

	return nil
}
