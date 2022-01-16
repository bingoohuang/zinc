package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prabhatsharma/zinc"
	"github.com/prabhatsharma/zinc/pkg/auth"
	"github.com/prabhatsharma/zinc/pkg/handlers"
	v1 "github.com/prabhatsharma/zinc/pkg/meta/v1"
)

// SetRoutes sets up all gi HTTP API endpoints that can be called by front end
func SetRoutes(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "authorization", "content-type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// meta service - health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/", func(c *gin.Context) { c.Redirect(301, "/ui/") })
	r.GET("/version", v1.GetVersion)
	r.StaticFS("/ui", http.FS(zinc.FrontendAssets))

	r.POST("/api/login", handlers.Login)
	r.PUT("/api/user", auth.ZincAuth, handlers.CreateUpdateUser)
	r.DELETE("/api/user/:userID", auth.ZincAuth, handlers.DeleteUser)
	r.GET("/api/users", auth.ZincAuth, handlers.GetUsers)

	r.PUT("/api/index", auth.ZincAuth, handlers.CreateIndex)
	r.GET("/api/index", auth.ZincAuth, handlers.ListIndexes)
	r.DELETE("/api/index/:indexName", auth.ZincAuth, handlers.DeleteIndex)

	// Bulk update/insert
	r.POST("/api/_bulk", auth.ZincAuth, handlers.BulkHandler)
	r.POST("/api/:target/_bulk", auth.ZincAuth, handlers.BulkHandler)

	// Document CRUD APIs. Update is same as create.
	r.PUT("/api/:target/doc", auth.ZincAuth, handlers.UpdateDoc)
	r.POST("/api/:target/_doc", auth.ZincAuth, handlers.UpdateDoc)
	r.PUT("/api/:target/_doc/:id", auth.ZincAuth, handlers.UpdateDoc)
	r.POST("/api/:target/_search", auth.ZincAuth, handlers.SearchIndex)
	r.DELETE("/api/:target/_doc/:id", auth.ZincAuth, handlers.DeleteDoc)
}
