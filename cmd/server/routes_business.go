package main

import (
	"github.com/labstack/echo/v4"
	"github.com/naozine/nz-magic-link/magiclink"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/appconfig"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/handlers"
)

// ConfigureBusinessSettings allows customization of MagicLink config and App Name
func ConfigureBusinessSettings(config *magiclink.Config) {
	config.RedirectURL = "/logistics/projects"         // Redirect to logistics projects list after login
	config.WebAuthnRedirectURL = "/logistics/projects" // Redirect to logistics projects list after passkey login

	// Set Application Name
	appconfig.AppName = "積込ログ管理"
}

// RegisterBusinessRoutes registers routes for business logic features
func RegisterBusinessRoutes(e *echo.Echo, queries *database.Queries, ml *magiclink.MagicLink) {
	// Handlers
	logisticsProjectHandler := handlers.NewLogisticsProjectHandler(queries)
	locationHandler := handlers.NewLocationHandler(queries)

	// Protected Routes (物流案件機能)
	logisticsGroup := e.Group("/logistics/projects")
	logisticsGroup.Use(ml.AuthMiddleware()) // Apply auth middleware to this group

	logisticsGroup.GET("", logisticsProjectHandler.ListLogisticsProjects)
	logisticsGroup.GET("/new", logisticsProjectHandler.NewLogisticsProjectPage)
	logisticsGroup.POST("/new", logisticsProjectHandler.CreateLogisticsProject)
	logisticsGroup.GET("/:id", logisticsProjectHandler.ShowLogisticsProject)
	logisticsGroup.GET("/:id/edit", logisticsProjectHandler.EditLogisticsProjectPage)
	logisticsGroup.POST("/:id/update", logisticsProjectHandler.UpdateLogisticsProject)
	logisticsGroup.POST("/:id/delete", logisticsProjectHandler.DeleteLogisticsProject)

	// Logistics Features (Course and Route Management) within a logistics project
	logisticsGroup.GET("/:id/courses/upload", logisticsProjectHandler.UploadRoutesPage)
	logisticsGroup.POST("/:id/courses/upload", logisticsProjectHandler.UploadRoutes)
	logisticsGroup.GET("/:id/courses", logisticsProjectHandler.ListCourses)
	logisticsGroup.GET("/:id/courses/:course_name", logisticsProjectHandler.ShowCourse)
	logisticsGroup.GET("/:id/courses/:course_name/location", logisticsProjectHandler.GetCurrentLocation) // htmx polling
	logisticsGroup.POST("/:id/courses/:course_name/reset", logisticsProjectHandler.ResetCourseStatus)
	logisticsGroup.GET("/:id/courses/:course_name/stops/:stop_id", logisticsProjectHandler.ShowStop)
	logisticsGroup.GET("/:id/courses/:course_name/stops/:stop_id/status", logisticsProjectHandler.GetStopTruckStatus) // htmx polling

	// API Routes (for external clients like mobile apps)
	apiGroup := e.Group("/api/v1")
	// Note: API authentication (e.g., API Key, Bearer Token) would typically be added here
	apiGroup.POST("/logistics/locations", locationHandler.CreateLocations)
}
