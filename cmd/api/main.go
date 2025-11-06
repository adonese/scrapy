package main

import (
	"log"

	"github.com/adonese/cost-of-living/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Basic middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", handlers.Health)

	// Start server
	log.Println("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
