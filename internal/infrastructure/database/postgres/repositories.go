package postgres

import (
	"database/sql"
	"tribute-back/internal/domain/entities"
	"tribute-back/internal/domain/repositories"

	"github.com/google/uuid"
)

type PgUserRepository struct {
	db *sql.DB
}

func NewPgUserRepository(db *sql.DB) repositories.UserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) FindByID(id int64) (*entities.User, error) {
	user := &entities.User{}
	// Note: The 'subscriptions' field is not in the 'users' table and will be populated in the service layer.
	query := `SELECT user_id, earned, is_verified, is_sub_published, is_onboarded, card_number FROM users WHERE user_id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Earned, &user.IsVerified, &user.IsSubPublished, &user.IsOnboarded, &user.CardNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a specific "not found" error
		}
		return nil, err
	}
	return user, nil
}

func (r *PgUserRepository) Update(user *entities.User) error {
	query := `UPDATE users SET earned = $2, is_verified = $3, is_sub_published = $4, is_onboarded = $5, card_number = $6 WHERE user_id = $1`
	_, err := r.db.Exec(query, user.ID, user.Earned, user.IsVerified, user.IsSubPublished, user.IsOnboarded, user.CardNumber)
	return err
}

func (r *PgUserRepository) Create(user *entities.User) error {
	query := `INSERT INTO users (user_id, earned, is_verified, is_sub_published, is_onboarded, card_number) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, user.ID, user.Earned, user.IsVerified, user.IsSubPublished, user.IsOnboarded, user.CardNumber)
	return err
}

// GetDB returns the underlying database connection
func (r *PgUserRepository) GetDB() *sql.DB {
	return r.db
}

type PgChannelRepository struct {
	db *sql.DB
}

func NewPgChannelRepository(db *sql.DB) repositories.ChannelRepository {
	return &PgChannelRepository{db: db}
}

func (r *PgChannelRepository) FindByUserID(userID int64) ([]*entities.Channel, error) {
	query := "SELECT id, user_id, channel_title, channel_username, is_verified FROM channels WHERE user_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*entities.Channel
	for rows.Next() {
		channel := &entities.Channel{}
		if err := rows.Scan(&channel.ID, &channel.UserID, &channel.ChannelTitle, &channel.ChannelUsername, &channel.IsVerified); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (r *PgChannelRepository) FindByID(id uuid.UUID) (*entities.Channel, error) {
	channel := &entities.Channel{}
	query := "SELECT id, user_id, channel_title, channel_username, is_verified FROM channels WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&channel.ID, &channel.UserID, &channel.ChannelTitle, &channel.ChannelUsername, &channel.IsVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return channel, nil
}

func (r *PgChannelRepository) Create(channel *entities.Channel) error {
	query := `INSERT INTO channels (id, user_id, channel_title, channel_username, is_verified) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, uuid.New(), channel.UserID, channel.ChannelTitle, channel.ChannelUsername, channel.IsVerified)
	return err
}

func (r *PgChannelRepository) Update(channel *entities.Channel) error {
	query := `UPDATE channels SET user_id = $2, channel_title = $3, channel_username = $4, is_verified = $5 WHERE id = $1`
	_, err := r.db.Exec(query, channel.ID, channel.UserID, channel.ChannelTitle, channel.ChannelUsername, channel.IsVerified)
	return err
}

func (r *PgChannelRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM channels WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

type PgSubscriptionRepository struct {
	db *sql.DB
}

func NewPgSubscriptionRepository(db *sql.DB) repositories.SubscriptionRepository {
	return &PgSubscriptionRepository{db: db}
}

func (r *PgSubscriptionRepository) FindByID(id uuid.UUID) (*entities.Subscription, error) {
	sub := &entities.Subscription{}
	query := `SELECT id, channel_id, user_id, channel_username, title, description, button_text, price, created_date FROM subscriptions WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&sub.ID, &sub.ChannelID, &sub.UserID, &sub.ChannelUsername, &sub.Title, &sub.Description, &sub.ButtonText, &sub.Price, &sub.CreatedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sub, nil
}

func (r *PgSubscriptionRepository) FindByUserID(userID int64) ([]*entities.Subscription, error) {
	query := `SELECT id, channel_id, user_id, channel_username, title, description, button_text, price, created_date FROM subscriptions WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*entities.Subscription
	for rows.Next() {
		sub := &entities.Subscription{}
		if err := rows.Scan(&sub.ID, &sub.ChannelID, &sub.UserID, &sub.ChannelUsername, &sub.Title, &sub.Description, &sub.ButtonText, &sub.Price, &sub.CreatedDate); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

func (r *PgSubscriptionRepository) Create(subscription *entities.Subscription) error {
	query := `INSERT INTO subscriptions (id, channel_id, user_id, channel_username, title, description, button_text, price, created_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query, uuid.New(), subscription.ChannelID, subscription.UserID, subscription.ChannelUsername, subscription.Title, subscription.Description, subscription.ButtonText, subscription.Price, subscription.CreatedDate)
	return err
}

func (r *PgSubscriptionRepository) Update(subscription *entities.Subscription) error {
	query := `UPDATE subscriptions SET title = $2, description = $3, button_text = $4, price = $5 WHERE id = $1`
	_, err := r.db.Exec(query, subscription.ID, subscription.Title, subscription.Description, subscription.ButtonText, subscription.Price)
	return err
}

func (r *PgSubscriptionRepository) FindByChannelID(channelID uuid.UUID) (*entities.Subscription, error) {
	sub := &entities.Subscription{}
	query := `SELECT id, channel_id, user_id, channel_username, title, description, button_text, price, created_date FROM subscriptions WHERE channel_id = $1`
	err := r.db.QueryRow(query, channelID).Scan(&sub.ID, &sub.ChannelID, &sub.UserID, &sub.ChannelUsername, &sub.Title, &sub.Description, &sub.ButtonText, &sub.Price, &sub.CreatedDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No subscription found for this channel, not an error
		}
		return nil, err
	}
	return sub, nil
}

type PgPaymentRepository struct {
	db *sql.DB
}

func NewPgPaymentRepository(db *sql.DB) repositories.PaymentRepository {
	return &PgPaymentRepository{db: db}
}

func (r *PgPaymentRepository) FindByUserID(userID int64) ([]*entities.Payment, error) {
	query := "SELECT id, user_id, description, created_date FROM payments WHERE user_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*entities.Payment
	for rows.Next() {
		payment := &entities.Payment{}
		if err := rows.Scan(&payment.ID, &payment.UserID, &payment.Description, &payment.CreatedDate); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

func (r *PgPaymentRepository) Create(payment *entities.Payment) error {
	query := `INSERT INTO payments (id, user_id, description, created_date) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, uuid.New(), payment.UserID, payment.Description, payment.CreatedDate)
	return err
}
