package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/velotrace/bikes-api/internal/handler"
	"velotrace.local/auth"
)

// @title VeloTrace Bikes API
// @version 1.0
// @description High-trust Bicycle Registry & Marketplace API.
// @host localhost:8081
// @BasePath /

func main() {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	origins := strings.Split(allowedOrigins, ",")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/identity?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse DATABASE_URL: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	bikeHandler := &handler.BikeHandler{DB: pool}

	// Public Routes
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	e.GET("/bikes", bikeHandler.ListBikesPublic)

	// Protected Routes
	jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
	protected := e.Group("")
	protected.Use(auth.JWTGuard(jwtPublicKey))

	protected.POST("/bikes", bikeHandler.RegisterBike)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
