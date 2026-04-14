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
	"github.com/velotrace/bikes-api/internal/handler"
	"github.com/velotrace/bikes-api/internal/platform"
	"github.com/velotrace/bikes-api/internal/repository"
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
	"velotrace.local/logger"
	"velotrace.local/utils"
)

// @title VeloTrace Bikes API
// @version 1.0
// @description High-trust Bicycle Registry & Marketplace API.
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	logger.Init("bikes-api")
	e := echo.New()
	e.Validator = utils.NewValidator()

	e.Use(middleware.RequestID())
	e.Use(logger.Middleware())
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.FromContext(c).Error("panic recovered",
				"err", err,
				"stack", string(stack),
			)
			return nil
		},
	}))

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

	// Storage Initialization
	storage, err := platform.NewStorage()
	if err != nil {
		logger.L.Error("failed to initialize storage", "err", err)
		os.Exit(1)
	}

	err = storage.VerifyConnection(context.Background())
	if err != nil {
		logger.L.Error("storage verification failed", "err", err)
		os.Exit(1)
	}
	storageBaseURL := os.Getenv("STORAGE_PUBLIC_BASE_URL")
	storageBucket := os.Getenv("STORAGE_BUCKET")
	if storageBaseURL == "" || storageBucket == "" {
		logger.L.Error("storage configuration missing",
			"missing_vars", []string{"STORAGE_PUBLIC_BASE_URL", "STORAGE_BUCKET"},
		)
		os.Exit(1)
	}

	logger.L.Info("storage connection successful and verified")

	// Wiring
	bikeRepo := repository.NewPgBikeRepository(pool)
	bikeService := service.NewBikeService(bikeRepo)
	bikeHandler := handler.NewBikeHandler(bikeService)

	imageRepo := repository.NewPgImageRepository(pool)
	imageService := service.NewImageService(imageRepo, bikeRepo, storage)
	imageHandler := handler.NewImageHandler(imageService)

	// Public Routes
	e.GET("/health", func(c echo.Context) error {
		l := logger.FromContext(c)
		err = pool.Ping(c.Request().Context())
		if err != nil {
			l.Error("Health check failed", "err", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "unhealthy"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	e.GET("/bikes", bikeHandler.ListMarketplace)

	// Protected Routes
	jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
	if jwtPublicKey == "" {
		logger.L.Error("environment configuration missing", "missing_var", "JWT_PUBLIC_KEY")
		os.Exit(1)
	}

	authManager, err := auth.NewTokenManager("", jwtPublicKey)
	if err != nil {
		logger.L.Error("failed to initialize auth manager", "err", err)
		os.Exit(1)
	}
	protected := e.Group("")
	protected.Use(authManager.JWTGuard())

	protected.POST("/bikes", bikeHandler.RegisterBike)
	protected.GET("/bikes/:id", bikeHandler.GetBike)
	protected.GET("/my/bikes", bikeHandler.ListMyBikes)
	protected.GET("/admin/bikes", bikeHandler.ListAdmin)

	protected.POST("/bikes/:id/upload-url", imageHandler.GetUploadURL)
	protected.POST("/bikes/:id/images/confirm", imageHandler.ConfirmUpload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
