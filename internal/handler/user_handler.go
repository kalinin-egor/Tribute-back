package handler

import (
	"net/http"
	"strconv"

	"tribute-back/internal/middleware"
	"tribute-back/internal/models"
	"tribute-back/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *service.UserService
	jwtConfig   middleware.Claims
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email, username, password, and personal information
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.CreateUserRequest true "User registration data"
// @Success      201  {object}  map[string]interface{}  "User created successfully"
// @Failure      400  {object}  map[string]interface{}  "Bad request - validation error"
// @Failure      409  {object}  map[string]interface{}  "Conflict - user already exists"
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password, return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Login credentials"
// @Success      200  {object}  models.LoginResponse  "Login successful"
// @Failure      400  {object}  map[string]interface{}  "Bad request - validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - invalid credentials"
// @Router       /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})

	tokenString, err := token.SignedString([]byte("your-secret-key")) // Use config
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		User:  *user,
		Token: tokenString,
	})
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Retrieve the profile of the currently authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "User profile"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - not authenticated"
// @Failure      404  {object}  map[string]interface{}  "Not found - user not found"
// @Router       /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Description  Update the profile information of the currently authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.UpdateUserRequest true "User update data"
// @Success      200  {object}  map[string]interface{}  "User updated successfully"
// @Failure      400  {object}  map[string]interface{}  "Bad request - validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - not authenticated"
// @Failure      404  {object}  map[string]interface{}  "Not found - user not found"
// @Router       /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieve a specific user by their ID (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}  "User data"
// @Failure      400  {object}  map[string]interface{}  "Bad request - invalid ID"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - not authenticated"
// @Failure      404  {object}  map[string]interface{}  "Not found - user not found"
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// ListUsers godoc
// @Summary      List all users
// @Description  Retrieve a paginated list of all users (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int  false  "Number of users to return (default: 10)"
// @Param        offset  query     int  false  "Number of users to skip (default: 0)"
// @Success      200     {object}  map[string]interface{}  "List of users"
// @Failure      400     {object}  map[string]interface{}  "Bad request - invalid parameters"
// @Failure      401     {object}  map[string]interface{}  "Unauthorized - not authenticated"
// @Router       /users/ [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	users, err := h.userService.ListUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Delete a specific user by their ID (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}  "User deleted successfully"
// @Failure      400  {object}  map[string]interface{}  "Bad request - invalid ID"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - not authenticated"
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
