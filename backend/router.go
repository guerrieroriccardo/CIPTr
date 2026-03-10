package main

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/handlers"
)

func setupRouter(database *sql.DB, jwtSecret []byte) *gin.Engine {
	r := gin.Default()

	// Trust only the loopback interface (reverse proxy runs on same host or Docker network).
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "172.16.0.0/12"})

	// CORS: permissive in development, tighten in production via env if needed.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CLI-Version"},
		AllowCredentials: false,
	}))

	authHandler := handlers.NewAuthHandler(database, jwtSecret)

	clientHandler       := handlers.NewClientHandler(database)
	siteHandler         := handlers.NewSiteHandler(database)
	addressBlockHandler := handlers.NewAddressBlockHandler(database)
	vlanHandler         := handlers.NewVLANHandler(database)
	locationHandler     := handlers.NewLocationHandler(database)
	manufacturerHandler := handlers.NewManufacturerHandler(database)
	categoryHandler     := handlers.NewCategoryHandler(database)
	supplierHandler     := handlers.NewSupplierHandler(database)
	osHandler           := handlers.NewOperatingSystemHandler(database)
	deviceModelHandler  := handlers.NewDeviceModelHandler(database)
	deviceHandler           := handlers.NewDeviceHandler(database)
	deviceInterfaceHandler  := handlers.NewDeviceInterfaceHandler(database)
	deviceIPHandler         := handlers.NewDeviceIPHandler(database)
	deviceConnectionHandler := handlers.NewDeviceConnectionHandler(database)
	switchHandler           := handlers.NewSwitchHandler(database)
	switchPortHandler       := handlers.NewSwitchPortHandler(database)
	patchPanelHandler       := handlers.NewPatchPanelHandler(database)
	patchPanelPortHandler   := handlers.NewPatchPanelPortHandler(database)
	deviceGroupHandler      := handlers.NewDeviceGroupHandler(database)
	deviceGroupMemberHandler := handlers.NewDeviceGroupMemberHandler(database)
	auditHandler            := handlers.NewAuditHandler(database)

	api := r.Group("/api/v1")
	api.Use(handlers.CLIVersionRequired())
	api.GET("/health", handlers.Health)
	api.POST("/login", authHandler.Login)
	api.POST("/guest-login", authHandler.GuestLogin)

	// All routes below require authentication.
	api.Use(handlers.AuthRequired(jwtSecret))
	api.GET("/me", authHandler.Me)
	api.PUT("/change-password", handlers.RoleRequired("viewer"), authHandler.ChangePassword)
	api.POST("/register", handlers.RoleRequired("admin"), authHandler.Register)
	api.GET("/users", handlers.RoleRequired("admin"), authHandler.ListUsers)
	api.PUT("/users/:id", handlers.RoleRequired("admin"), authHandler.UpdateUser)

	// Shorthand for technician-level write protection.
	write := handlers.RoleRequired("technician")

	clients := api.Group("/clients")
	{
		clients.GET("", clientHandler.List)
		clients.POST("", write, clientHandler.Create)
		clients.GET("/:id", clientHandler.GetByID)
		clients.PUT("/:id", write, clientHandler.Update)
		clients.DELETE("/:id", write, clientHandler.Delete)
		clients.GET("/:id/sites", siteHandler.ListByClient)
	}

	sites := api.Group("/sites")
	{
		sites.GET("", siteHandler.List)
		sites.POST("", write, siteHandler.Create)
		sites.GET("/:id", siteHandler.GetByID)
		sites.PUT("/:id", write, siteHandler.Update)
		sites.DELETE("/:id", write, siteHandler.Delete)
		sites.GET("/:id/address-blocks", addressBlockHandler.ListBySite)
		sites.GET("/:id/vlans", vlanHandler.ListBySite)
		sites.GET("/:id/locations", locationHandler.ListBySite)
		sites.GET("/:id/devices", deviceHandler.ListBySite)
		sites.GET("/:id/switches", switchHandler.ListBySite)
		sites.GET("/:id/patch-panels", patchPanelHandler.ListBySite)
		sites.GET("/:id/device-groups", deviceGroupHandler.ListBySite)
	}

	manufacturers := api.Group("/manufacturers")
	{
		manufacturers.GET("", manufacturerHandler.List)
		manufacturers.POST("", write, manufacturerHandler.Create)
		manufacturers.GET("/:id", manufacturerHandler.GetByID)
		manufacturers.PUT("/:id", write, manufacturerHandler.Update)
		manufacturers.DELETE("/:id", write, manufacturerHandler.Delete)
	}

	categoriesGroup := api.Group("/categories")
	{
		categoriesGroup.GET("", categoryHandler.List)
		categoriesGroup.POST("", write, categoryHandler.Create)
		categoriesGroup.GET("/:id", categoryHandler.GetByID)
		categoriesGroup.PUT("/:id", write, categoryHandler.Update)
		categoriesGroup.DELETE("/:id", write, categoryHandler.Delete)
	}

	operatingSystems := api.Group("/operating-systems")
	{
		operatingSystems.GET("", osHandler.List)
		operatingSystems.POST("", write, osHandler.Create)
		operatingSystems.GET("/:id", osHandler.GetByID)
		operatingSystems.PUT("/:id", write, osHandler.Update)
		operatingSystems.DELETE("/:id", write, osHandler.Delete)
	}

	suppliersGroup := api.Group("/suppliers")
	{
		suppliersGroup.GET("", supplierHandler.List)
		suppliersGroup.POST("", write, supplierHandler.Create)
		suppliersGroup.GET("/:id", supplierHandler.GetByID)
		suppliersGroup.PUT("/:id", write, supplierHandler.Update)
		suppliersGroup.DELETE("/:id", write, supplierHandler.Delete)
	}

	patchPanels := api.Group("/patch-panels")
	{
		patchPanels.GET("", patchPanelHandler.List)
		patchPanels.POST("", write, patchPanelHandler.Create)
		patchPanels.GET("/:id", patchPanelHandler.GetByID)
		patchPanels.PUT("/:id", write, patchPanelHandler.Update)
		patchPanels.DELETE("/:id", write, patchPanelHandler.Delete)
		patchPanels.GET("/:id/ports", patchPanelPortHandler.ListByPatchPanel)
	}

	patchPanelPorts := api.Group("/patch-panel-ports")
	{
		patchPanelPorts.GET("", patchPanelPortHandler.List)
		patchPanelPorts.POST("", write, patchPanelPortHandler.Create)
		patchPanelPorts.GET("/:id", patchPanelPortHandler.GetByID)
		patchPanelPorts.PUT("/:id", write, patchPanelPortHandler.Update)
		patchPanelPorts.DELETE("/:id", write, patchPanelPortHandler.Delete)
	}

	switches := api.Group("/switches")
	{
		switches.GET("", switchHandler.List)
		switches.GET("/next-name", switchHandler.NextName)
		switches.POST("", write, switchHandler.Create)
		switches.GET("/:id", switchHandler.GetByID)
		switches.PUT("/:id", write, switchHandler.Update)
		switches.DELETE("/:id", write, switchHandler.Delete)
		switches.GET("/:id/ports", switchPortHandler.ListBySwitch)
	}

	switchPorts := api.Group("/switch-ports")
	{
		switchPorts.GET("", switchPortHandler.List)
		switchPorts.POST("", write, switchPortHandler.Create)
		switchPorts.GET("/:id", switchPortHandler.GetByID)
		switchPorts.PUT("/:id", write, switchPortHandler.Update)
		switchPorts.DELETE("/:id", write, switchPortHandler.Delete)
	}

	addressBlocks := api.Group("/address-blocks")
	{
		addressBlocks.GET("", addressBlockHandler.List)
		addressBlocks.POST("", write, addressBlockHandler.Create)
		addressBlocks.GET("/:id", addressBlockHandler.GetByID)
		addressBlocks.PUT("/:id", write, addressBlockHandler.Update)
		addressBlocks.DELETE("/:id", write, addressBlockHandler.Delete)
		addressBlocks.GET("/:id/vlans", vlanHandler.ListByAddressBlock)
	}

	locations := api.Group("/locations")
	{
		locations.GET("", locationHandler.List)
		locations.POST("", write, locationHandler.Create)
		locations.GET("/:id", locationHandler.GetByID)
		locations.PUT("/:id", write, locationHandler.Update)
		locations.DELETE("/:id", write, locationHandler.Delete)
	}

	devices := api.Group("/devices")
	{
		devices.GET("", deviceHandler.List)
		devices.GET("/next-hostname", deviceHandler.NextHostname)
		devices.POST("", write, deviceHandler.Create)
		devices.GET("/:id", deviceHandler.GetByID)
		devices.PUT("/:id", write, deviceHandler.Update)
		devices.DELETE("/:id", write, deviceHandler.Delete)
		devices.GET("/:id/interfaces", deviceInterfaceHandler.ListByDevice)
		devices.GET("/:id/ips", deviceIPHandler.ListByDevice)
		devices.GET("/:id/connections", deviceConnectionHandler.ListByDevice)
		devices.GET("/:id/label", deviceHandler.Label)
	}

	deviceInterfaces := api.Group("/device-interfaces")
	{
		deviceInterfaces.GET("", deviceInterfaceHandler.List)
		deviceInterfaces.POST("", write, deviceInterfaceHandler.Create)
		deviceInterfaces.GET("/:id", deviceInterfaceHandler.GetByID)
		deviceInterfaces.PUT("/:id", write, deviceInterfaceHandler.Update)
		deviceInterfaces.DELETE("/:id", write, deviceInterfaceHandler.Delete)
	}

	deviceIPs := api.Group("/device-ips")
	{
		deviceIPs.GET("", deviceIPHandler.List)
		deviceIPs.POST("", write, deviceIPHandler.Create)
		deviceIPs.GET("/:id", deviceIPHandler.GetByID)
		deviceIPs.PUT("/:id", write, deviceIPHandler.Update)
		deviceIPs.DELETE("/:id", write, deviceIPHandler.Delete)
	}

	deviceConnections := api.Group("/device-connections")
	{
		deviceConnections.GET("", deviceConnectionHandler.List)
		deviceConnections.POST("", write, deviceConnectionHandler.Create)
		deviceConnections.GET("/:id", deviceConnectionHandler.GetByID)
		deviceConnections.PUT("/:id", write, deviceConnectionHandler.Update)
		deviceConnections.DELETE("/:id", write, deviceConnectionHandler.Delete)
	}

	deviceModels := api.Group("/device-models")
	{
		deviceModels.GET("", deviceModelHandler.List)
		deviceModels.POST("", write, deviceModelHandler.Create)
		deviceModels.GET("/:id", deviceModelHandler.GetByID)
		deviceModels.PUT("/:id", write, deviceModelHandler.Update)
		deviceModels.DELETE("/:id", write, deviceModelHandler.Delete)
	}

	vlans := api.Group("/vlans")
	{
		vlans.GET("", vlanHandler.List)
		vlans.POST("", write, vlanHandler.Create)
		vlans.GET("/:id", vlanHandler.GetByID)
		vlans.PUT("/:id", write, vlanHandler.Update)
		vlans.DELETE("/:id", write, vlanHandler.Delete)
	}

	deviceGroups := api.Group("/device-groups")
	{
		deviceGroups.GET("", deviceGroupHandler.List)
		deviceGroups.POST("", write, deviceGroupHandler.Create)
		deviceGroups.GET("/:id", deviceGroupHandler.GetByID)
		deviceGroups.PUT("/:id", write, deviceGroupHandler.Update)
		deviceGroups.DELETE("/:id", write, deviceGroupHandler.Delete)
		deviceGroups.GET("/:id/members", deviceGroupMemberHandler.ListByGroup)
	}

	deviceGroupMembers := api.Group("/device-group-members")
	{
		deviceGroupMembers.GET("", deviceGroupMemberHandler.List)
		deviceGroupMembers.POST("", write, deviceGroupMemberHandler.Create)
		deviceGroupMembers.DELETE("/:id", write, deviceGroupMemberHandler.Delete)
	}

	api.GET("/audit-logs", handlers.RoleRequired("admin"), auditHandler.List)

	return r
}
