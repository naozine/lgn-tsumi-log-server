package main

import (
	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-magic-link/magiclink"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/appconfig"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/handlers"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/mdm"
	appMiddleware "github.com/naozine/project_crud_with_auth_tmpl/internal/middleware"
)

// ConfigureBusinessSettings allows customization of MagicLink config and App Name
func ConfigureBusinessSettings(config *magiclink.Config) {
	config.RedirectURL = "/projects"         // Redirect to projects list after login
	config.WebAuthnRedirectURL = "/projects" // Redirect to projects list after passkey login

	// Set Application Name
	appconfig.AppName = "積込ログ管理"
}

// RegisterBusinessRoutes registers routes for business logic features
func RegisterBusinessRoutes(e *echo.Echo, queries *database.Queries, ml *magiclink.MagicLink, mdmClient *mdm.Client) {
	// Handlers
	projectHandler := handlers.NewProjectHandler(queries)
	locationHandler := handlers.NewLocationHandler(queries)
	mdmHandler := handlers.NewMDMHandler(mdmClient)

	// Protected Routes (物流案件機能 - projectsとして上書き)
	projectGroup := e.Group("/projects")
	projectGroup.Use(appMiddleware.RequireAuth(ml, "/auth/login")) // 未認証時はログインページへリダイレクト

	projectGroup.GET("", projectHandler.ListProjects)
	projectGroup.GET("/new", projectHandler.NewProjectPage)
	projectGroup.POST("/new", projectHandler.CreateProject)
	projectGroup.GET("/:id", projectHandler.ShowProject)
	projectGroup.GET("/:id/edit", projectHandler.EditProjectPage)
	projectGroup.POST("/:id/update", projectHandler.UpdateProject)
	projectGroup.POST("/:id/delete", projectHandler.DeleteProject)
	projectGroup.POST("/:id/api-key", projectHandler.RegenerateAPIKey)

	// Logistics Features (Course and Route Management) within a logistics project
	projectGroup.GET("/:id/courses/upload", projectHandler.UploadRoutesPage)
	projectGroup.POST("/:id/courses/upload", projectHandler.UploadRoutes)
	projectGroup.GET("/:id/courses", projectHandler.ListCourses)
	projectGroup.GET("/:id/courses/:course_name", projectHandler.ShowCourse)
	projectGroup.GET("/:id/courses/:course_name/location", projectHandler.GetCurrentLocation) // htmx polling
	projectGroup.POST("/:id/courses/:course_name/reset", projectHandler.ResetCourseStatus)
	projectGroup.GET("/:id/courses/:course_name/stops/:stop_id", projectHandler.ShowStop)
	projectGroup.GET("/:id/courses/:course_name/stops/:stop_id/status", projectHandler.GetStopTruckStatus) // htmx polling

	// Device Management
	projectGroup.POST("/:id/devices/:device_id/assign", projectHandler.AssignDeviceCourse)
	projectGroup.POST("/:id/devices/:device_id/delete", projectHandler.DeleteDevice)

	// API Routes (for external clients like mobile apps)
	apiGroup := e.Group("/api/v1")
	// Note: API authentication (e.g., API Key, Bearer Token) would typically be added here
	apiGroup.POST("/devices", locationHandler.RegisterDevice)
	apiGroup.POST("/locations", locationHandler.CreateLocations)
	apiGroup.POST("/photos", locationHandler.CreatePhotoMetadata)
	apiGroup.POST("/photos/upload", locationHandler.UploadPhoto)

	// MDM Routes (admin only)
	mdmGroup := e.Group("/mdm")
	mdmGroup.Use(appMiddleware.RequireAuth(ml, "/auth/login"))
	mdmGroup.Use(appMiddleware.RequireRole("admin"))
	mdmGroup.GET("", mdmHandler.MDMTop)
	mdmGroup.GET("/devices", mdmHandler.ListMDMDevices)
	mdmGroup.GET("/devices/:id", mdmHandler.ShowMDMDevice)
	mdmGroup.GET("/apps", mdmHandler.ListMDMApps)
}
