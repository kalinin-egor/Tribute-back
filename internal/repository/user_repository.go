package repository

import (
	"database/sql"

	"tribute-back/internal/models"

	"github.com/google/uuid"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(query,
		user.ID, user.Email, user.Username, user.Password,
		user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt)

	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, username, password, first_name, last_name, created_at, updated_at
		FROM users WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, username, password, first_name, last_name, created_at, updated_at
		FROM users WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, username, password, first_name, last_name, created_at, updated_at
		FROM users WHERE username = $1
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, username = $3, first_name = $4, last_name = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.Exec(query,
		user.ID, user.Email, user.Username,
		user.FirstName, user.LastName, user.UpdatedAt)

	return err
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves all users with pagination
func (r *UserRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.Password,
			&user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
