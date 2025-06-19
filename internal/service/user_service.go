package service

import (
	"errors"
	"time"

	"tribute-back/internal/models"
	"tribute-back/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic for users
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUser creates a new user with hashed password
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if email is already taken by another user
		if existingUser, err := s.userRepo.GetByEmail(*req.Email); err == nil && existingUser.ID != id {
			return nil, errors.New("email already taken")
		}
		user.Email = *req.Email
	}

	if req.Username != nil {
		// Check if username is already taken by another user
		if existingUser, err := s.userRepo.GetByUsername(*req.Username); err == nil && existingUser.ID != id {
			return nil, errors.New("username already taken")
		}
		user.Username = *req.Username
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}

// ListUsers retrieves users with pagination
func (s *UserService) ListUsers(limit, offset int) ([]*models.User, error) {
	return s.userRepo.List(limit, offset)
}

// AuthenticateUser authenticates a user with email and password
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
