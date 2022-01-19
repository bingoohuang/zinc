package main

import (
	"errors"
	"fmt"
	"github.com/prabhatsharma/zinc"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/bingoohuang/golog"
	"github.com/prabhatsharma/zinc/pkg/routes"
	"github.com/prabhatsharma/zinc/pkg/zutil"
)

func init() {
	golog.Setup()
	zinc.Init()
}

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Print("Error loading .env file")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	routes.SetRoutes(r)

	port := zutil.GetEnvInt("PORT", 4080)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Start to run on address: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Printf("Run failed: %v", err)
	}
}
