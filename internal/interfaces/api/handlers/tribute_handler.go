package handlers

import (
	"net/http"
	"strings"
	"time"
	"tribute-back/internal/application/services"
	"tribute-back/internal/infrastructure/payouts"
	"tribute-back/internal/interfaces/api/dto"

	"github.com/gin-gonic/gin"
)

type TributeHandler struct {
	service *services.TributeService
}

func NewTributeHandler(service *services.TributeService) *TributeHandler {
	return &TributeHandler{service: service}
}

// buildDashboardResponse creates a dashboard response from dashboard data
func (h *TributeHandler) buildDashboardResponse(data *services.DashboardData) *dto.DashboardResponse {
	return &dto.DashboardResponse{
		Earn:           data.User.Earned,
		IsVerified:     data.User.IsVerified,
		IsSubPublished: data.User.IsSubPublished,
		CardNumber:     data.User.CardNumber,
		ChannelsAndGroups: func() []dto.ChannelDTO {
			dtos := make([]dto.ChannelDTO, len(data.Channels))
			for i, ch := range data.Channels {
				dtos[i] = dto.ChannelDTO{ID: ch.ID, ChannelTitle: ch.ChannelTitle, ChannelUsername: ch.ChannelUsername, IsVerified: ch.IsVerified}
			}
			return dtos
		}(),
		Subscriptions: func() []dto.SubDTO {
			dtos := make([]dto.SubDTO, len(data.Subscriptions))
			for i, sub := range data.Subscriptions {
				dtos[i] = dto.SubDTO{ID: sub.ID, Title: sub.Title, Description: sub.Description, Price: sub.Price}
			}
			return dtos
		}(),
		PaymentsHistory: func() []dto.PaymentDTO {
			dtos := make([]dto.PaymentDTO, len(data.Payments))
			for i, p := range data.Payments {
				dtos[i] = dto.PaymentDTO{Description: p.Description, CreatedDate: p.CreatedDate.Format(time.RFC3339)}
			}
			return dtos
		}(),
	}
}

// This function is no longer needed as routes are registered directly in server.go
// You can remove it or leave it empty.
func (h *TributeHandler) RegisterRoutes(api *gin.RouterGroup) {
	// Routes are now registered in server.go to separate public and private endpoints.
}

// @Summary      Get Dashboard Data
// @Description  Retrieves all data for the main dashboard screen. The user is identified via the `initData` in the Authorization header. If the user does not exist in the database, a 404 error is returned.
// @Tags         Tribute
// @Produce      json
// @Security     TgAuth
// @Success      200  {object}  dto.DashboardResponse  "Successfully retrieved dashboard data."
// @Failure      401  {object}  dto.ErrorResponse      "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse      "Forbidden - The provided initData is invalid or expired."
// @Failure      404  {object}  dto.ErrorResponse      "Not Found - The user does not exist in the database."
// @Failure      500  {object}  dto.ErrorResponse      "Internal Server Error - An unexpected error occurred."
// @Router       /dashboard [get]
func (h *TributeHandler) Dashboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}
	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}
	data, err := h.service.GetDashboardData(id)
	if err != nil {
		if err.Error() == "user not found" {
			// Return 404 error when user doesn't exist
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	response := h.buildDashboardResponse(data)
	c.JSON(http.StatusOK, response)
}

// @Summary      Onboard a User
// @Description  Creates a user record if one doesn't exist, or updates an existing user to mark them as onboarded. This is the first endpoint a new user should call. It is idempotent.
// @Tags         Tribute
// @Produce      json
// @Security     TgAuth
// @Success      200  {object}  dto.OnboardResponse  "Success - The user already existed and has been marked as onboarded."
// @Success      201  {object}  dto.OnboardResponse  "Created - A new user was created and marked as onboarded."
// @Failure      401  {object}  dto.ErrorResponse    "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse    "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse    "Internal Server Error - An unexpected error occurred."
// @Router       /onboard [put]
func (h *TributeHandler) Onboard(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User ID not found in context"})
		return
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "User ID has an invalid type"})
		return
	}

	user, created, err := h.service.OnboardUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.OnboardResponse{
		Message: "User is onboarded successfully",
		User: dto.UserResponse{
			ID:             user.ID,
			Earned:         user.Earned,
			IsVerified:     user.IsVerified,
			IsSubPublished: user.IsSubPublished,
			IsOnboarded:    user.IsOnboarded,
			CardNumber:     user.CardNumber,
		},
	}

	if created {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// @Summary      Add a new Channel
// @Description  Adds a new Telegram channel for the specified user. The channel is saved with is_verified = false. User must exist in the system.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Param        payload body dto.AddBotRequest true "The user ID, channel title and username to add."
// @Success      201  {object}  dto.AddBotResponse     "Created - The channel was added successfully."
// @Failure      400  {object}  dto.ErrorResponse      "Bad Request - The request body is invalid, user not found, or channel is already added."
// @Failure      500  {object}  dto.ErrorResponse      "Internal Server Error - Database error."
// @Router       /add-bot [post]
func (h *TributeHandler) AddBot(c *gin.Context) {
	var req dto.AddBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Check if user exists
	_, err := h.service.GetDashboardData(req.UserID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	channel, err := h.service.AddBot(req.UserID, req.ChannelTitle, req.ChannelUsername)
	if err != nil {
		// Check if it's a business logic error (channel already exists)
		if err.Error() == "this channel is already added to your account" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			return
		}
		// For other errors (database, etc.) return 500
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.AddBotResponse{
		Message: "Channel added successfully",
		Channel: dto.ChannelDTO{
			ID:              channel.ID,
			ChannelTitle:    channel.ChannelTitle,
			ChannelUsername: channel.ChannelUsername,
			IsVerified:      channel.IsVerified,
		},
	})
}

// @Summary      Upload Documents for Verification
// @Description  Uploads a user's photo and passport scan for manual verification. Both images must be provided as base64 encoded strings. The documents are sent to a private admin chat for review.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     TgAuth
// @Param        payload body dto.UploadVerifiedPassportRequest true "JSON object containing base64 encoded photo and passport."
// @Success      200  {object}  dto.MessageResponse    "Success - The verification request was sent successfully."
// @Failure      400  {object}  dto.ErrorResponse      "Bad Request - The request body is invalid or missing required fields."
// @Failure      401  {object}  dto.ErrorResponse      "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse      "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse      "Internal Server Error - Failed to send documents to the verification service."
// @Router       /upload-verified-passport [post]
func (h *TributeHandler) UploadVerifiedPassport(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}
	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	var req dto.UploadVerifiedPassportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	if req.UserPhoto == "" || req.UserPassport == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "user-photo and user-passport are required"})
		return
	}

	err := h.service.RequestVerification(id, req.UserPhoto, req.UserPassport)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to send verification request: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Verification request sent successfully"})
}

// @Summary      Telegram Verification Webhook
// @Description  **PUBLIC ENDPOINT.** This endpoint is intended to be called by Telegram in response to an admin clicking a button in the verification chat. It should not be called directly by the frontend. It processes verification approvals and rejections.
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        payload body dto.TelegramUpdate true "The callback query update sent by Telegram."
// @Success      200  {object}  dto.StatusResponse     "Success - The callback was processed."
// @Failure      400  {object}  dto.ErrorResponse      "Bad Request - The payload from Telegram is malformed."
// @Failure      500  {object}  dto.ErrorResponse      "Internal Server Error - Failed to process the callback data."
// @Router       /check-verified-passport [post]
func (h *TributeHandler) CheckVerifiedPassport(c *gin.Context) {
	var update dto.TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Cannot parse Telegram update"})
		return
	}

	if update.CallbackQuery == nil || update.CallbackQuery.Message == nil {
		c.JSON(http.StatusOK, dto.StatusResponse{Status: "Not a valid callback query, ignoring"})
		return
	}

	err := h.service.HandleVerificationCallback(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		update.CallbackQuery.Data,
	)

	if err != nil {
		// In a real app, you might want to answer the callback query with an error message to the admin.
		// For now, just log it and return an error.
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to process callback: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.StatusResponse{Status: "ok"})
}

// @Summary      Set Up Payout Method
// @Description  Registers a user's bank card as a payout method. **IMPORTANT:** Card details are NOT stored in our database. They are forwarded directly to a secure payment gateway. The user must be verified to use this endpoint.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     TgAuth
// @Param        payload body dto.SetUpPayoutsRequest true "The user's card details."
// @Success      200  {object}  dto.MessageResponse    "Success - The payout method was registered successfully."
// @Failure      400  {object}  dto.ErrorResponse      "Bad Request - The request body is invalid."
// @Failure      401  {object}  dto.ErrorResponse      "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse      "Forbidden - The provided initData is invalid or expired, or the user is not verified."
// @Failure      500  {object}  dto.ErrorResponse      "Internal Server Error - The payment gateway returned an error."
// @Router       /set-up-payouts [post]
func (h *TributeHandler) SetUpPayouts(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User ID not found in context"})
		return
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "User ID has an invalid type"})
		return
	}

	var req dto.SetUpPayoutsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	cardDetails := payouts.CardDetails{
		CardNumber: req.CardNumber,
		CardDate:   req.CardDate,
		CardCVV:    req.CardCVV,
	}

	if err := h.service.SetUpPayouts(id, cardDetails); err != nil {
		if strings.Contains(err.Error(), "user must be verified") {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to set up payouts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Payout method set up successfully"})
}

// @Summary      Publish or Update a Subscription Tier
// @Description  Allows an author to create or update their public subscription details (title, description, price). This is an idempotent operation. The user must have at least one channel added via `/add-bot` to use this.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     TgAuth
// @Param        payload body dto.PublishSubscriptionRequest true "The details of the subscription tier to publish."
// @Success      200  {object}  dto.PublishSubscriptionResponse "Success - The subscription was published or updated successfully."
// @Failure      400  {object}  dto.ErrorResponse               "Bad Request - The request body is invalid."
// @Failure      401  {object}  dto.ErrorResponse               "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse               "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse               "Internal Server Error - e.g., the user has no channels."
// @Router       /publish-subscription [put]
func (h *TributeHandler) PublishSubscription(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}
	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	var req dto.PublishSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	subscription, err := h.service.PublishSubscription(id, req.Title, req.Description, req.ButtonText, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PublishSubscriptionResponse{
		Message: "Subscription published successfully",
		Subscription: dto.SubDTO{
			ID:          subscription.ID,
			Title:       subscription.Title,
			Description: subscription.Description,
			Price:       subscription.Price,
		},
	})
}

// @Summary      Subscribe to an Author
// @Description  Allows an authenticated user (the subscriber) to pay for and subscribe to another user (the creator). This action creates a payment record and updates the creator's earnings.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     TgAuth
// @Param        payload body dto.CreateSubscribeRequest true "The ID of the user to subscribe to and the price."
// @Success      201  {object}  dto.MessageResponse      "Created - The subscription was successful."
// @Failure      400  {object}  dto.ErrorResponse        "Bad Request - The request body is invalid."
// @Failure      401  {object}  dto.ErrorResponse        "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse        "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse        "Internal Server Error - e.g., the creator has no subscription tier, or the price is incorrect."
// @Router       /create-subscribe [post]
func (h *TributeHandler) CreateSubscribe(c *gin.Context) {
	subscriberID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}
	id, ok := subscriberID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	var req dto.CreateSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	// The user making the request is the subscriber. The user_id in the body is the creator.
	if err := h.service.CreateSubscription(id, req.UserID, req.Price); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MessageResponse{Message: "Successfully subscribed"})
}

// @Summary      Create User
// @Description  Creates a new user if one doesn't exist, otherwise returns the existing user. This endpoint is idempotent and returns dashboard data.
// @Tags         Tribute
// @Produce      json
// @Security     TgAuth
// @Success      200  {object}  dto.DashboardResponse  "Success - User already existed."
// @Success      201  {object}  dto.DashboardResponse  "Created - A new user was created."
// @Failure      401  {object}  dto.ErrorResponse       "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse       "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse       "Internal Server Error - An unexpected error occurred."
// @Router       /create-user [post]
func (h *TributeHandler) CreateUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	_, err := h.service.CreateUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Get dashboard data to return in response
	data, err := h.service.GetDashboardData(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := h.buildDashboardResponse(data)

	// Check if user was created or already existed
	existingUser, _ := h.service.GetDashboardData(id)
	created := existingUser == nil || existingUser.User == nil

	if created {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// @Summary      Reset Database
// @Description  Drops all tables and recreates them with empty structure. WARNING: This will delete all data!
// @Tags         Development
// @Produce      json
// @Success      200  {object}  dto.MessageResponse  "Success - Database was reset successfully."
// @Failure      500  {object}  dto.ErrorResponse    "Internal Server Error - An unexpected error occurred."
// @Router       /reset-database [get]
func (h *TributeHandler) ResetDatabase(c *gin.Context) {
	if err := h.service.ResetDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Database reset successfully"})
}

// @Summary      Get Channel List
// @Description  Returns a list of all channels for the authenticated user.
// @Tags         Tribute
// @Produce      json
// @Security     TgAuth
// @Success      200  {array}   dto.ChannelDTO          "Success - List of user's channels."
// @Failure      401  {object}  dto.ErrorResponse       "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse       "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse       "Internal Server Error - Database error."
// @Router       /channel-list [get]
func (h *TributeHandler) GetChannelList(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	channels, err := h.service.GetChannelList(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert to DTO
	channelDTOs := make([]dto.ChannelDTO, len(channels))
	for i, ch := range channels {
		channelDTOs[i] = dto.ChannelDTO{
			ID:              ch.ID,
			ChannelTitle:    ch.ChannelTitle,
			ChannelUsername: ch.ChannelUsername,
			IsVerified:      ch.IsVerified,
		}
	}

	c.JSON(http.StatusOK, channelDTOs)
}

// @Summary      Check Channel Ownership
// @Description  Checks if the user is the owner of the specified channel. If yes, sets is_verified = true. If no, deletes the channel.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     TgAuth
// @Param        payload body dto.CheckChannelRequest true "The channel ID to check."
// @Success      200  {object}  dto.CheckChannelResponse "Success - Channel ownership check result."
// @Failure      400  {object}  dto.ErrorResponse        "Bad Request - The request body is invalid or channel not found."
// @Failure      401  {object}  dto.ErrorResponse        "Unauthorized - The Authorization header is missing or invalid."
// @Failure      403  {object}  dto.ErrorResponse        "Forbidden - The provided initData is invalid or expired."
// @Failure      500  {object}  dto.ErrorResponse        "Internal Server Error - Failed to verify channel ownership."
// @Router       /check-channel [post]
func (h *TributeHandler) CheckChannel(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not authenticated"})
		return
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	var req dto.CheckChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	isOwner, err := h.service.CheckChannel(id, req.ChannelID)
	if err != nil {
		// Check if it's a business logic error (channel not found, not owned by user)
		if err.Error() == "channel not found" || err.Error() == "channel does not belong to this user" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			return
		}
		// For other errors (network, database, etc.) return 500
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.CheckChannelResponse{
		IsOwner: isOwner,
	})
}
