package handlers

import (
	"net/http"
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

// This function is no longer needed as routes are registered directly in server.go
// You can remove it or leave it empty.
func (h *TributeHandler) RegisterRoutes(api *gin.RouterGroup) {
	// Routes are now registered in server.go to separate public and private endpoints.
}

// @Summary      Get Dashboard Data
// @Description  Retrieves all data for the main dashboard screen based on the authenticated user.
// @Tags         Tribute
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  dto.DashboardResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
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
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "User not found. Please complete onboarding."})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	response := &dto.DashboardResponse{
		Earn:           data.User.Earned,
		IsVerified:     data.User.IsVerified,
		IsSubPublished: data.User.IsSubPublished,
		ChannelsAndGroups: func() []dto.ChannelDTO {
			dtos := make([]dto.ChannelDTO, len(data.Channels))
			for i, ch := range data.Channels {
				dtos[i] = dto.ChannelDTO{ID: ch.ID, ChannelUsername: ch.ChannelUsername}
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
	c.JSON(http.StatusOK, response)
}

// @Summary      Add a new bot/channel
// @Description  Adds a new Telegram channel for the authenticated user.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.AddBotRequest true "Bot Username"
// @Success      201  {object}  dto.AddBotResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /add-bot [post]
func (h *TributeHandler) AddBot(c *gin.Context) {
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

	var req dto.AddBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	channel, err := h.service.AddBot(id, req.BotUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.AddBotResponse{
		Message: "Bot added successfully",
		Channel: dto.ChannelDTO{
			ID:              channel.ID,
			ChannelUsername: channel.ChannelUsername,
		},
	})
}

// @Summary      Upload documents for verification
// @Description  Uploads user photo and passport (as base64 strings) to start the verification process.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.UploadVerifiedPassportRequest true "User Documents"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
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

// @Summary      Telegram Webhook for Verification
// @Description  This is a public endpoint for receiving callback queries from Telegram for verification approval or rejection.
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        update body dto.TelegramUpdate true "Telegram Callback Query"
// @Success      200  {object}  dto.StatusResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
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

// @Summary      Set up payout method
// @Description  Sets up the payout method for the authenticated user by providing card details. The details are not stored.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.SetUpPayoutsRequest true "Card Details"
// @Success      200  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /set-up-payouts [post]
func (h *TributeHandler) SetUpPayouts(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Payout method set up successfully"})
}

// @Summary      Publish or update a subscription tier
// @Description  Allows an author to publish or update their subscription details.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.PublishSubscriptionRequest true "Subscription Details"
// @Success      200  {object}  dto.PublishSubscriptionResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
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

// @Summary      Create a subscription to an author
// @Description  Allows an authenticated user (subscriber) to subscribe to another user (creator).
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body dto.CreateSubscribeRequest true "Subscription Request"
// @Success      201  {object}  dto.MessageResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
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

// @Summary      Onboard a new user
// @Description  Creates or updates a user record. Marks the user as onboarded.
// @Tags         Tribute
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  dto.OnboardResponse "User already existed and was updated"
// @Success      201  {object}  dto.OnboardResponse "User created successfully"
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /onboard [put]
func (h *TributeHandler) Onboard(c *gin.Context) {
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

	user, err := h.service.OnboardUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// A real implementation would distinguish between create and update to set status code
	c.JSON(http.StatusOK, dto.OnboardResponse{
		Message: "User is onboarded",
		User: dto.UserResponse{
			ID:             user.ID,
			Earned:         user.Earned,
			IsVerified:     user.IsVerified,
			IsSubPublished: user.IsSubPublished,
			IsOnboarded:    user.IsOnboarded,
		},
	})
}
