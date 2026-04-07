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
	"github.com/velotrace/bikes-api/internal/platform"
	"github.com/velotrace/bikes-api/internal/repository"
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
	"velotrace.local/utils"
)

// @title VeloTrace Bikes API
// @version 1.0
// @description High-trust Bicycle Registry & Marketplace API.
// @host localhost:8081
// @BasePath /

func main() {
	e := echo.New()
	e.Validator = utils.NewValidator()

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

	// Storage Initialization
	storage, err := platform.NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	err = storage.VerifyConnection(context.Background())
	if err != nil {
		log.Fatalf("Storage verification failed: %v", err)
	}
	storageBaseURL := os.Getenv("STORAGE_PUBLIC_BASE_URL")
	storageBucket := os.Getenv("STORAGE_BUCKET")
	if storageBaseURL == "" || storageBucket == "" {
		log.Fatalf("STORAGE_PUBLIC_BASE_URL and STORAGE_BUCKET must be set")
	}

	log.Println("Storage connection successful and verified")

	// Wiring
	bikeRepo := repository.NewPgBikeRepository(pool)
	bikeService := service.NewBikeService(bikeRepo)
	bikeHandler := handler.NewBikeHandler(bikeService)

	imageRepo := repository.NewPgImageRepository(pool)
	imageService := service.NewImageService(imageRepo, bikeRepo, storage)
	imageHandler := handler.NewImageHandler(imageService)

	// Public Routes
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	e.GET("/bikes", bikeHandler.ListMarketplace)

	// Protected Routes
	jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
	protected := e.Group("")
	protected.Use(auth.JWTGuard(jwtPublicKey))

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
