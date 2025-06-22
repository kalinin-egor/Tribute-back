package server

import (
	"database/sql"
	"log"
	"net/http"
	"tribute-back/internal/application/services"
	"tribute-back/internal/config"
	"tribute-back/internal/infrastructure/auth"
	"tribute-back/internal/infrastructure/database/postgres"
	"tribute-back/internal/infrastructure/payouts"
	"tribute-back/internal/infrastructure/telegram"
	"tribute-back/internal/interfaces/api/handlers"
	"tribute-back/internal/interfaces/api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewServer(db *sql.DB, redisClient *redis.Client) *gin.Engine {
	router := gin.Default()

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Replace with your allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Infrastructure Services
	tgAuthService, err := auth.NewTelegramAuthService()
	if err != nil {
		log.Fatal("Failed to initialize Telegram Auth Service: ", err)
	}
	botService, err := telegram.NewBotService()
	if err != nil {
		log.Fatal("Failed to initialize Telegram Bot Service:", err)
	}
	payoutGateway := payouts.NewMockGateway()

	// Repositories
	userRepo := postgres.NewPgUserRepository(db)
	channelRepo := postgres.NewPgChannelRepository(db)
	subRepo := postgres.NewPgSubscriptionRepository(db)
	paymentRepo := postgres.NewPgPaymentRepository(db)

	// Application Services
	tributeService := services.NewTributeService(userRepo, channelRepo, subRepo, paymentRepo, botService, payoutGateway)

	// Handlers
	tributeHandler := handlers.NewTributeHandler(tributeService)

	// Public webhook for Telegram
	router.POST("/api/v1/check-verified-passport", tributeHandler.CheckVerifiedPassport)

	// Protected routes
	api := router.Group("/api/v1")
	api.Use(middleware.TelegramAuthMiddleware(tgAuthService))
	{
		api.GET("/dashboard", tributeHandler.Dashboard)
		api.PUT("/onboard", tributeHandler.Onboard)
		api.POST("/create-user", tributeHandler.CreateUser)
		api.POST("/add-bot", tributeHandler.AddBot)
		api.POST("/upload-verified-passport", tributeHandler.UploadVerifiedPassport)
		api.POST("/set-up-payouts", tributeHandler.SetUpPayouts)
		api.PUT("/publish-subscription", tributeHandler.PublishSubscription)
		api.POST("/create-subscribe", tributeHandler.CreateSubscribe)
	}

	// Swagger - no test routes needed anymore
	if config.GetEnv("GIN_MODE", "debug") != "release" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}
