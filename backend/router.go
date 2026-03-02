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

	clientHandler      := handlers.NewClientHandler(database)
	siteHandler        := handlers.NewSiteHandler(database)
	addressBlockHandler := handlers.NewAddressBlockHandler(database)
	vlanHandler        := handlers.NewVLANHandler(database)
	locationHandler    := handlers.NewLocationHandler(database)
	deviceModelHandler := handlers.NewDeviceModelHandler(database)
	deviceHandler          := handlers.NewDeviceHandler(database)
	deviceInterfaceHandler := handlers.NewDeviceInterfaceHandler(database)

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
		api.GET("/sites/:id/address-blocks", addressBlockHandler.ListBySite)
		api.GET("/sites/:id/vlans", vlanHandler.ListBySite)
		api.GET("/sites/:id/locations", locationHandler.ListBySite)
		api.GET("/sites/:id/devices", deviceHandler.ListBySite)

		// Address Blocks
		api.GET("/address-blocks", addressBlockHandler.List)
		api.POST("/address-blocks", addressBlockHandler.Create)
		api.GET("/address-blocks/:id", addressBlockHandler.GetByID)
		api.PUT("/address-blocks/:id", addressBlockHandler.Update)
		api.DELETE("/address-blocks/:id", addressBlockHandler.Delete)
		api.GET("/address-blocks/:id/vlans", vlanHandler.ListByAddressBlock)

		// Locations
		api.GET("/locations", locationHandler.List)
		api.POST("/locations", locationHandler.Create)
		api.GET("/locations/:id", locationHandler.GetByID)
		api.PUT("/locations/:id", locationHandler.Update)
		api.DELETE("/locations/:id", locationHandler.Delete)

		// Devices
		api.GET("/devices", deviceHandler.List)
		api.POST("/devices", deviceHandler.Create)
		api.GET("/devices/:id", deviceHandler.GetByID)
		api.PUT("/devices/:id", deviceHandler.Update)
		api.DELETE("/devices/:id", deviceHandler.Delete)
		api.GET("/devices/:id/interfaces", deviceInterfaceHandler.ListByDevice)

		// Device Interfaces
		api.GET("/device-interfaces", deviceInterfaceHandler.List)
		api.POST("/device-interfaces", deviceInterfaceHandler.Create)
		api.GET("/device-interfaces/:id", deviceInterfaceHandler.GetByID)
		api.PUT("/device-interfaces/:id", deviceInterfaceHandler.Update)
		api.DELETE("/device-interfaces/:id", deviceInterfaceHandler.Delete)

		// Device Models
		api.GET("/device-models", deviceModelHandler.List)
		api.POST("/device-models", deviceModelHandler.Create)
		api.GET("/device-models/:id", deviceModelHandler.GetByID)
		api.PUT("/device-models/:id", deviceModelHandler.Update)
		api.DELETE("/device-models/:id", deviceModelHandler.Delete)

		// VLANs
		api.GET("/vlans", vlanHandler.List)
		api.POST("/vlans", vlanHandler.Create)
		api.GET("/vlans/:id", vlanHandler.GetByID)
		api.PUT("/vlans/:id", vlanHandler.Update)
		api.DELETE("/vlans/:id", vlanHandler.Delete)
	}

	return r
}
