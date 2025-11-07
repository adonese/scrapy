package main

import (
	"log"
	"os"

	"github.com/adonese/cost-of-living/internal/handlers"
	customMiddleware "github.com/adonese/cost-of-living/internal/middleware"
	"github.com/adonese/cost-of-living/internal/repository/postgres"
	"github.com/adonese/cost-of-living/internal/services/estimator"
	uihandlers "github.com/adonese/cost-of-living/internal/ui/handlers"
	"github.com/adonese/cost-of-living/pkg/database"
	"github.com/adonese/cost-of-living/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize structured logger
	logger.Init()
	logger.Info("Starting UAE Cost of Living API")

	// Connect to database
	cfg := database.NewConfigFromEnv()
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	costDataPointRepo := postgres.NewCostDataPointRepository(db.GetConn())
	logger.Info("Initialized CostDataPointRepository")

	// Aggregation/estimator service
	estimatorService := estimator.NewService(costDataPointRepo, nil)

	// Initialize Echo
	e := echo.New()

	// Basic middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(customMiddleware.ErrorHandler())
	e.Use(customMiddleware.MetricsMiddleware())

	// Static assets
	e.Static("/static", "web/static")

	// Public UI routes
	homeHandler := uihandlers.NewHomeHandler(estimatorService)
	e.GET("/", homeHandler.Index)
	e.POST("/ui/estimate", homeHandler.EstimatePartial)

	// Health check route
	e.GET("/health", handlers.NewHealthHandler(db).Health)

	// Metrics endpoint for Prometheus
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// API v1 routes
	api := e.Group("/api/v1")

	// Cost data points endpoints
	costDataPointHandler := handlers.NewCostDataPointHandler(costDataPointRepo)
	api.POST("/cost-data-points", costDataPointHandler.Create)
	api.GET("/cost-data-points/:id", costDataPointHandler.GetByID)
	api.GET("/cost-data-points", costDataPointHandler.List)
	api.PUT("/cost-data-points/:id", costDataPointHandler.Update)
	api.DELETE("/cost-data-points/:id", costDataPointHandler.Delete)

	// Estimates + summary endpoints
	estimateHandler := handlers.NewEstimatorHandler(estimatorService)
	api.POST("/estimates", estimateHandler.Estimate)
	api.GET("/estimates/summary", estimateHandler.Summary)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", "port", port)
	if err := e.Start(":" + port); err != nil {
		logger.Error("Server failed to start", "error", err)
		log.Fatal(err)
	}
}
