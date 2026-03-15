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
	"github.com/velotrace/identity-service/internal/handler"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
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

	userHandler := &handler.UserHandler{DB: pool}

	e.GET("/health", func(c echo.Context) error {
		err := pool.Ping(context.Background())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "VeloTrace Identity Service")
	})

	e.POST("/auth/google", userHandler.AuthGoogle)

	port := os.Getenv("PORT")

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
