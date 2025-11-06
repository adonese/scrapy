package main

import (
	"log"
	"os"

	"github.com/adonese/cost-of-living/internal/handlers"
	customMiddleware "github.com/adonese/cost-of-living/internal/middleware"
	"github.com/adonese/cost-of-living/internal/repository/postgres"
	"github.com/adonese/cost-of-living/pkg/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Connect to database
	cfg := database.NewConfigFromEnv()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	costDataPointRepo := postgres.NewCostDataPointRepository(db.GetConn())
	log.Printf("Initialized CostDataPointRepository")

	// Initialize Echo
	e := echo.New()

	// Basic middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(customMiddleware.ErrorHandler())

	// Health check route
	e.GET("/health", handlers.NewHealthHandler(db).Health)

	// API v1 routes
	api := e.Group("/api/v1")

	// Cost data points endpoints
	costDataPointHandler := handlers.NewCostDataPointHandler(costDataPointRepo)
	api.POST("/cost-data-points", costDataPointHandler.Create)
	api.GET("/cost-data-points/:id", costDataPointHandler.GetByID)
	api.GET("/cost-data-points", costDataPointHandler.List)
	api.PUT("/cost-data-points/:id", costDataPointHandler.Update)
	api.DELETE("/cost-data-points/:id", costDataPointHandler.Delete)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
