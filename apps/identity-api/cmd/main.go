package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/velotrace/identity-api/internal/handler"
	"github.com/velotrace/identity-api/internal/repository"
	"github.com/velotrace/identity-api/internal/service"
	"velotrace.local/auth"
	"velotrace.local/logger"
)

// @title VeloTrace Identity API
// @version 1.0
// @description High-trust Bicycle Registry Identity Service.
// @host localhost:8080
// @BasePath /

func main() {
	logger.Init("identity-api")
	e := echo.New()

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(logger.Middleware())

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
		logger.L.Error("unable to parse DATABASE_URL", "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.L.Error("unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	privateKey := os.Getenv("JWT_PRIVATE_KEY")
	publicKey := os.Getenv("JWT_PUBLIC_KEY")
	if privateKey == "" || publicKey == "" {
		logger.L.Error("environment configuration missing", "missing_var", "JWT_PUBLIC_KEY", "missing_var", "JWT_PRIVATE_KEY")
		os.Exit(1)
	}

	authManager, err := auth.NewTokenManager(privateKey, publicKey)
	if err != nil {
		logger.L.Error("failed to initialize auth manager", "err", err)
		os.Exit(1)
	}

	userRepo := repository.NewPgUserRepository(pool)
	userService := service.NewUserService(userRepo, authManager)
	userHandler := handler.NewUserHandler(userService)

	e.GET("/health", func(c echo.Context) error {
		l := logger.FromContext(c)
		err := pool.Ping(c.Request().Context())
		if err != nil {
			l.Error("Health check failed", "err", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "unhealthy"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "VeloTrace Identity Service")
	})

	e.POST("/auth/google", userHandler.AuthGoogle)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
