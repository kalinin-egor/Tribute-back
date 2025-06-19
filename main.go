package main

import (
	"log"
	"os"

	"tribute-back/internal/config"
	"tribute-back/internal/database"
	"tribute-back/internal/redis"
	"tribute-back/internal/server"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Error loading environment variables:", err)
	}

	// Initialize database
	db, err := database.Init()
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.Init()
	if err != nil {
		log.Fatal("Error initializing Redis:", err)
	}
	defer redisClient.Close()

	// Initialize and start server
	app := server.NewServer(db, redisClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
