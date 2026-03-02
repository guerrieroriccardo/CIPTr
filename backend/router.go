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
	deviceIPHandler         := handlers.NewDeviceIPHandler(database)
	deviceConnectionHandler := handlers.NewDeviceConnectionHandler(database)
	switchHandler           := handlers.NewSwitchHandler(database)
	switchPortHandler       := handlers.NewSwitchPortHandler(database)

	api := r.Group("/api/v1")
	api.GET("/health", handlers.Health)

	clients := api.Group("/clients")
	{
		clients.GET("", clientHandler.List)
		clients.POST("", clientHandler.Create)
		clients.GET("/:id", clientHandler.GetByID)
		clients.PUT("/:id", clientHandler.Update)
		clients.DELETE("/:id", clientHandler.Delete)
		clients.GET("/:id/sites", siteHandler.ListByClient)
	}

	sites := api.Group("/sites")
	{
		sites.GET("", siteHandler.List)
		sites.POST("", siteHandler.Create)
		sites.GET("/:id", siteHandler.GetByID)
		sites.PUT("/:id", siteHandler.Update)
		sites.DELETE("/:id", siteHandler.Delete)
		sites.GET("/:id/address-blocks", addressBlockHandler.ListBySite)
		sites.GET("/:id/vlans", vlanHandler.ListBySite)
		sites.GET("/:id/locations", locationHandler.ListBySite)
		sites.GET("/:id/devices", deviceHandler.ListBySite)
		sites.GET("/:id/switches", switchHandler.ListBySite)
	}

	switches := api.Group("/switches")
	{
		switches.GET("", switchHandler.List)
		switches.POST("", switchHandler.Create)
		switches.GET("/:id", switchHandler.GetByID)
		switches.PUT("/:id", switchHandler.Update)
		switches.DELETE("/:id", switchHandler.Delete)
		switches.GET("/:id/ports", switchPortHandler.ListBySwitch)
	}

	switchPorts := api.Group("/switch-ports")
	{
		switchPorts.GET("", switchPortHandler.List)
		switchPorts.POST("", switchPortHandler.Create)
		switchPorts.GET("/:id", switchPortHandler.GetByID)
		switchPorts.PUT("/:id", switchPortHandler.Update)
		switchPorts.DELETE("/:id", switchPortHandler.Delete)
	}

	addressBlocks := api.Group("/address-blocks")
	{
		addressBlocks.GET("", addressBlockHandler.List)
		addressBlocks.POST("", addressBlockHandler.Create)
		addressBlocks.GET("/:id", addressBlockHandler.GetByID)
		addressBlocks.PUT("/:id", addressBlockHandler.Update)
		addressBlocks.DELETE("/:id", addressBlockHandler.Delete)
		addressBlocks.GET("/:id/vlans", vlanHandler.ListByAddressBlock)
	}

	locations := api.Group("/locations")
	{
		locations.GET("", locationHandler.List)
		locations.POST("", locationHandler.Create)
		locations.GET("/:id", locationHandler.GetByID)
		locations.PUT("/:id", locationHandler.Update)
		locations.DELETE("/:id", locationHandler.Delete)
	}

	devices := api.Group("/devices")
	{
		devices.GET("", deviceHandler.List)
		devices.POST("", deviceHandler.Create)
		devices.GET("/:id", deviceHandler.GetByID)
		devices.PUT("/:id", deviceHandler.Update)
		devices.DELETE("/:id", deviceHandler.Delete)
		devices.GET("/:id/interfaces", deviceInterfaceHandler.ListByDevice)
		devices.GET("/:id/ips", deviceIPHandler.ListByDevice)
		devices.GET("/:id/connections", deviceConnectionHandler.ListByDevice)
	}

	deviceInterfaces := api.Group("/device-interfaces")
	{
		deviceInterfaces.GET("", deviceInterfaceHandler.List)
		deviceInterfaces.POST("", deviceInterfaceHandler.Create)
		deviceInterfaces.GET("/:id", deviceInterfaceHandler.GetByID)
		deviceInterfaces.PUT("/:id", deviceInterfaceHandler.Update)
		deviceInterfaces.DELETE("/:id", deviceInterfaceHandler.Delete)
	}

	deviceIPs := api.Group("/device-ips")
	{
		deviceIPs.GET("", deviceIPHandler.List)
		deviceIPs.POST("", deviceIPHandler.Create)
		deviceIPs.GET("/:id", deviceIPHandler.GetByID)
		deviceIPs.PUT("/:id", deviceIPHandler.Update)
		deviceIPs.DELETE("/:id", deviceIPHandler.Delete)
	}

	deviceConnections := api.Group("/device-connections")
	{
		deviceConnections.GET("", deviceConnectionHandler.List)
		deviceConnections.POST("", deviceConnectionHandler.Create)
		deviceConnections.GET("/:id", deviceConnectionHandler.GetByID)
		deviceConnections.PUT("/:id", deviceConnectionHandler.Update)
		deviceConnections.DELETE("/:id", deviceConnectionHandler.Delete)
	}

	deviceModels := api.Group("/device-models")
	{
		deviceModels.GET("", deviceModelHandler.List)
		deviceModels.POST("", deviceModelHandler.Create)
		deviceModels.GET("/:id", deviceModelHandler.GetByID)
		deviceModels.PUT("/:id", deviceModelHandler.Update)
		deviceModels.DELETE("/:id", deviceModelHandler.Delete)
	}

	vlans := api.Group("/vlans")
	{
		vlans.GET("", vlanHandler.List)
		vlans.POST("", vlanHandler.Create)
		vlans.GET("/:id", vlanHandler.GetByID)
		vlans.PUT("/:id", vlanHandler.Update)
		vlans.DELETE("/:id", vlanHandler.Delete)
	}

	return r
}
