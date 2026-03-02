package main

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ciptr/handlers"
)

func setupRouter(database *sql.DB) *gin.Engine {
	r := gin.Default()

	// Trust only the loopback interface (reverse proxy runs on same host or Docker network).
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "172.16.0.0/12"})

	// CORS: permissive in development, tighten in production via env if needed.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
	}))

	clientHandler := handlers.NewClientHandler(database)
	siteHandler := handlers.NewSiteHandler(database)

	api := r.Group("/api/v1")
	{
		api.GET("/health", handlers.Health)

		// Clients
		api.GET("/clients", clientHandler.List)
		api.POST("/clients", clientHandler.Create)
		api.GET("/clients/:id", clientHandler.GetByID)
		api.PUT("/clients/:id", clientHandler.Update)
		api.DELETE("/clients/:id", clientHandler.Delete)
		api.GET("/clients/:id/sites", siteHandler.ListByClient)

		// Sites
		api.GET("/sites", siteHandler.List)
		api.POST("/sites", siteHandler.Create)
		api.GET("/sites/:id", siteHandler.GetByID)
		api.PUT("/sites/:id", siteHandler.Update)
		api.DELETE("/sites/:id", siteHandler.Delete)
	}

	return r
}
