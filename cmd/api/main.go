package main

import (
	"log"
	"os"

	"github.com/adonese/cost-of-living/internal/handlers"
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

	// Routes
	e.GET("/health", handlers.NewHealthHandler(db).Health)

	// TODO: Add API handlers for cost data points in iteration 1.4
	_ = costDataPointRepo // Will be used in next iteration

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
