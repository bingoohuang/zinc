package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/prabhatsharma/zinc/pkg/routes"
	"github.com/prabhatsharma/zinc/pkg/zutils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("Error loading .env file")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	routes.SetRoutes(r) // Set up all API routes.

	// Run the server
	port := zutils.GetEnv("PORT", "4080")
	if err := r.Run(":" + port); err != nil {
		log.Printf("Run failed: %v", err)
	}
}
