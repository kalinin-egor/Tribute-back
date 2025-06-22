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

func (s *TributeService) AddBot(userID int64, botUsername string) (*entities.Channel, error) {
	// Send initial notification
	initialMessage := fmt.Sprintf("Just a moment, we are checking bot permissions in %s", botUsername)
	if err := s.telegramBot.SendMessage(userID, initialMessage); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send initial message to user %d: %v\n", userID, err)
	}

	// Check if user is owner/admin of the channel
	chatMember, err := s.telegramBot.CheckChannelMembership(botUsername, userID)
	if err != nil {
		// Send error message to user
		errorMessage := "Failed to check channel permissions. Please try again."
		s.telegramBot.SendMessage(userID, errorMessage)
		return nil, fmt.Errorf("failed to check channel membership: %w", err)
	}

	// Only allow owners and administrators to add their channels
	if chatMember.Status != "creator" && chatMember.Status != "administrator" {
		// Send rejection message to user
		rejectionMessage := "You are not the owner of this channel."
		s.telegramBot.SendMessage(userID, rejectionMessage)
		return nil, errors.New("you must be the owner or administrator of this channel to add it")
	}

	// Optional: Check if the bot (channel) already exists for this user to prevent duplicates
	existingChannels, err := s.channels.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, ch := range existingChannels {
		if ch.ChannelUsername == botUsername {
			// Send duplicate message to user
			duplicateMessage := fmt.Sprintf("Channel %s is already added to your account.", botUsername)
			s.telegramBot.SendMessage(userID, duplicateMessage)
			return nil, errors.New("this channel is already added to your account")
		}
	}

	channel := &entities.Channel{
		UserID:          userID,
		ChannelUsername: botUsername,
	}

	err = s.channels.Create(channel)
	if err != nil {
		return nil, err
	}

	// Send success message to user
	successMessage := fmt.Sprintf("Good! You added bot to channel: %s (@%s)", botUsername, botUsername)
	if err := s.telegramBot.SendMessage(userID, successMessage); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send success message to user %d: %v\n", userID, err)
	}

	return channel, nil
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

func (s *TributeService) SetUpPayouts(userID int64, cardDetails payouts.CardDetails) error {
	// Here you could add any business logic before contacting the payment gateway.
	// For example, check if the user is verified.
	user, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	if !user.IsVerified {
		return errors.New("user must be verified to set up payouts")
	}

	return s.payoutGateway.RegisterPayoutMethod(userID, cardDetails)
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

	// Finally, update the user's status to indicate a sub is published
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	user.IsSubPublished = true
	if err := s.users.Update(user); err != nil {
		// Log error but don't fail the whole operation
	}

	return subscription, nil
}

func (s *TributeService) CreateSubscription(subscriberID int64, creatorID int64, price float64) error {
	// Find the creator's channel and subscription tier
	channels, err := s.channels.FindByUserID(creatorID)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		return errors.New("creator has no channels")
	}
	channel := channels[0] // Assuming first channel

	subscription, err := s.subs.FindByChannelID(channel.ID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return errors.New("creator has not published a subscription tier")
	}

	// Validate the price
	if subscription.Price != price {
		return fmt.Errorf("incorrect price provided. expected %.2f, got %.2f", subscription.Price, price)
	}

	// Create payment record
	payment := &entities.Payment{
		UserID:      creatorID, // The creator receives the payment
		Description: fmt.Sprintf("Subscription payment from user %d", subscriberID),
		CreatedDate: time.Now(),
	}
	if err := s.payments.Create(payment); err != nil {
		return fmt.Errorf("failed to create payment record: %w", err)
	}

	// Update creator's earnings
	creator, err := s.users.FindByID(creatorID)
	if err != nil {
		return fmt.Errorf("failed to find creator to update earnings: %w", err)
	}
	creator.Earned += price
	if err := s.users.Update(creator); err != nil {
		return fmt.Errorf("failed to update creator earnings: %w", err)
	}

	return nil
}

func (s *TributeService) OnboardUser(userID int64) (user *entities.User, created bool, err error) {
	// Check if user already exists
	existingUser, err := s.users.FindByID(userID)
	if err != nil && err.Error() != "user not found" { // A real error occurred
		return nil, false, err
	}
	if existingUser != nil {
		if !existingUser.IsOnboarded {
			existingUser.IsOnboarded = true
			if err := s.users.Update(existingUser); err != nil {
				return nil, false, err
			}
			return existingUser, false, nil // User existed, but was now onboarded (updated)
		}
		return existingUser, false, nil // User already existed and was onboarded
	}

	// Create new user if not found
	newUser := &entities.User{
		ID:          userID,
		Earned:      0,
		IsVerified:  false,
		IsOnboarded: true, // Mark as onboarded on creation
	}

	if err := s.users.Create(newUser); err != nil {
		return nil, false, err
	}

	return newUser, true, nil
}

// CreateUser creates a new user if one doesn't exist, otherwise returns existing user
func (s *TributeService) CreateUser(userID int64) (*entities.User, error) {
	// Check if user already exists
	existingUser, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		// User already exists, return it without creating
		return existingUser, nil
	}

	// Create new user if not found
	newUser := &entities.User{
		ID:          userID,
		Earned:      0,
		IsVerified:  false,
		IsOnboarded: false,
	}

	if err := s.users.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// ResetDatabase drops all tables and recreates them with empty structure
func (s *TributeService) ResetDatabase() error {
	// Get the database connection from the repository
	// We need to access the raw DB connection to execute DDL statements
	userRepo, ok := s.users.(*postgres.PgUserRepository)
	if !ok {
		return errors.New("failed to get database connection")
	}

	db := userRepo.GetDB()

	// Drop all tables in correct order (due to foreign key constraints)
	dropQueries := []string{
		"DROP TABLE IF EXISTS payments CASCADE",
		"DROP TABLE IF EXISTS subscriptions CASCADE",
		"DROP TABLE IF EXISTS channels CASCADE",
		"DROP TABLE IF EXISTS users CASCADE",
	}

	for _, query := range dropQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	// Recreate all tables
	createQueries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id BIGINT PRIMARY KEY,
			earned NUMERIC(10, 2) DEFAULT 0.00,
			is_verified BOOLEAN DEFAULT FALSE,
			is_sub_published BOOLEAN DEFAULT FALSE,
			is_onboarded BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS channels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			channel_username VARCHAR(255) UNIQUE NOT NULL
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

	for _, query := range createQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// TODO: Implement methods for each endpoint
// e.g. GetDashboardData, AddBot, UploadPassport, etc.
