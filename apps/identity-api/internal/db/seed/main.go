//go:build devseed

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/velotrace/identity-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/identity?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	fmt.Println("🌱 Seeding Users...")

	// 1. Create specific Mock/Test Users for predictable testing
	testUsers := []models.User{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Email:       "tester@velotrace.local",
			GoogleID:    "mock-google-id-1",
			DisplayName: "Standard Tester",
			IsVerified:  true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:       "verified-owner@velotrace.local",
			GoogleID:    "mock-google-id-2",
			DisplayName: "Verified Owner",
			IsVerified:  true,
		},
	}

	for _, u := range testUsers {
		err := db.Where(models.User{Email: u.Email}).FirstOrCreate(&u).Error
		if err != nil {
			log.Printf("Could not seed test user %s: %v", u.Email, err)
		} else {
			fmt.Printf("Mock user ready: %s\n", u.Email)
		}
	}

	// 2. Add 10 random users using Faker
	for i := 0; i < 10; i++ {
		firstName := faker.FirstName()
		lastName := faker.LastName()
		user := models.User{
			ID:          uuid.New(),
			Email:       faker.Email(),
			GoogleID:    faker.UUIDDigit(),
			DisplayName: fmt.Sprintf("%s %s", firstName, lastName),
			FirstName:   &firstName,
			LastName:    &lastName,
			IsVerified:  true,
		}

		result := db.Create(&user)
		if result.Error != nil {
			log.Printf("Could not seed faker user: %v", result.Error)
			continue
		}
	}

	fmt.Println("✅ User seeding and mocking completed!")
}
