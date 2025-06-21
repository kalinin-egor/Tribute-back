package service

import (
	"database/sql"
	"testing"

	"tribute-back/internal/models"
	"tribute-back/internal/repository"

	"github.com/google/uuid"
)

// MockUserRepository is a mock implementation of the user repository for testing
type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() repository.UserRepositoryInterface {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	m.users[user.ID.String()] = user
	return nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	if user, exists := m.users[id.String()]; exists {
		return user, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) Update(user *models.User) error {
	m.users[user.ID.String()] = user
	return nil
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	delete(m.users, id.String())
	return nil
}

func (m *MockUserRepository) List(limit, offset int) ([]*models.User, error) {
	users := make([]*models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

// TestUserService_CreateUser tests the CreateUser method
func TestUserService_CreateUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	user, err := userService.CreateUser(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Error("Expected user to be created, got nil")
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}

	if user.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, user.Username)
	}

	if user.FirstName != req.FirstName {
		t.Errorf("Expected first name %s, got %s", req.FirstName, user.FirstName)
	}

	if user.LastName != req.LastName {
		t.Errorf("Expected last name %s, got %s", req.LastName, user.LastName)
	}
}

// TestUserService_CreateUser_DuplicateEmail tests creating a user with duplicate email
func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	// Create first user
	req1 := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser1",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User1",
	}

	_, err := userService.CreateUser(req1)
	if err != nil {
		t.Errorf("Expected no error for first user, got %v", err)
	}

	// Try to create second user with same email
	req2 := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser2",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User2",
	}

	_, err = userService.CreateUser(req2)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

// TestUserService_CreateUser_DuplicateUsername tests creating a user with duplicate username
func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	// Create first user
	req1 := &models.CreateUserRequest{
		Email:     "test1@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User1",
	}

	_, err := userService.CreateUser(req1)
	if err != nil {
		t.Errorf("Expected no error for first user, got %v", err)
	}

	// Try to create second user with same username
	req2 := &models.CreateUserRequest{
		Email:     "test2@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User2",
	}

	_, err = userService.CreateUser(req2)
	if err == nil {
		t.Error("Expected error for duplicate username, got nil")
	}
}

// TestUserService_GetUserByID tests the GetUserByID method
func TestUserService_GetUserByID(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	// Create a test user first
	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := userService.CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test getting the user by ID
	retrievedUser, err := userService.GetUserByID(createdUser.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedUser == nil {
		t.Error("Expected user to be retrieved, got nil")
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %s, got %s", createdUser.ID, retrievedUser.ID)
	}
}

// TestUserService_AuthenticateUser tests the AuthenticateUser method
func TestUserService_AuthenticateUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	// Create a test user
	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	_, err := userService.CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test successful authentication
	user, err := userService.AuthenticateUser("test@example.com", "password123")
	if err != nil {
		t.Errorf("Expected no error for valid credentials, got %v", err)
	}

	if user == nil {
		t.Error("Expected user to be returned, got nil")
	}

	// Test failed authentication with wrong password
	_, err = userService.AuthenticateUser("test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for invalid password, got nil")
	}

	// Test failed authentication with non-existent email
	_, err = userService.AuthenticateUser("nonexistent@example.com", "password123")
	if err == nil {
		t.Error("Expected error for non-existent email, got nil")
	}
}

// TestUserService_UpdateUser tests the UpdateUser method
func TestUserService_UpdateUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	// Create a test user
	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := userService.CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update the user
	newFirstName := "Updated"
	newLastName := "Name"
	updateReq := &models.UpdateUserRequest{
		FirstName: &newFirstName,
		LastName:  &newLastName,
	}

	updatedUser, err := userService.UpdateUser(createdUser.ID, updateReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if updatedUser.FirstName != newFirstName {
		t.Errorf("Expected first name %s, got %s", newFirstName, updatedUser.FirstName)
	}

	if updatedUser.LastName != newLastName {
		t.Errorf("Expected last name %s, got %s", newLastName, updatedUser.LastName)
	}
}

// TestUserService_DeleteUser tests the DeleteUser method
func TestUserService_DeleteUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userService := NewUserService(mockRepo)

	// Create a test user
	req := &models.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := userService.CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete the user
	err = userService.DeleteUser(createdUser.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify user is deleted
	_, err = userService.GetUserByID(createdUser.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user, got nil")
	}
}
