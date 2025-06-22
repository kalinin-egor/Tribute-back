package services

import (
	"tribute-back/internal/domain/entities"

	"github.com/google/uuid"
)

// UserService defines the interface for user business operations
type UserService interface {
	CreateUser(email, username, password, firstName, lastName string) (*entities.User, error)
	GetUserByID(id uuid.UUID) (*entities.User, error)
	GetUserByEmail(email string) (*entities.User, error)
	UpdateUser(id uuid.UUID, email, username, firstName, lastName string) (*entities.User, error)
	DeleteUser(id uuid.UUID) error
	ListUsers(limit, offset int) ([]*entities.User, error)
	AuthenticateUser(email, password string) (*entities.User, error)
}
