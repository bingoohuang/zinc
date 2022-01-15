package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"

	"github.com/prabhatsharma/zinc/pkg/routes"
	"github.com/prabhatsharma/zinc/pkg/zutil"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Print("Error loading .env file")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	routes.SetRoutes(r) // Set up all API routes.

	// Run the server
	port := zutil.GetEnv("PORT", "4080")
	if err := r.Run(":" + port); err != nil {
		log.Printf("Run failed: %v", err)
	}
}
