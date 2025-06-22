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
	// Optional: Check if the bot (channel) already exists for this user to prevent duplicates
	existingChannels, err := s.channels.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, ch := range existingChannels {
		if ch.ChannelUsername == botUsername {
			return nil, errors.New("bot with this username already exists for the user")
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

// TODO: Implement methods for each endpoint
// e.g. GetDashboardData, AddBot, UploadPassport, etc.
