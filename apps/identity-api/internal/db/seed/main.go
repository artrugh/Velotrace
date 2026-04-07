//go:build devseed

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID          uuid.UUID
	Email       string
	GoogleID    string
	DisplayName string
	FirstName   *string
	LastName    *string
	IsVerified  bool
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/identity?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("🌱 Seeding Users...")

	// 1. Create specific Mock/Test Users for predictable testing
	testUsers := []User{
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
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, email, google_id, display_name, is_verified)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (email) DO NOTHING
		`, u.ID, u.Email, u.GoogleID, u.DisplayName, u.IsVerified)

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
		email := faker.Email()
		googleID := faker.UUIDDigit()
		displayName := fmt.Sprintf("%s %s", firstName, lastName)

		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, email, google_id, display_name, first_name, last_name, is_verified)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email) DO NOTHING
		`, uuid.New(), email, googleID, displayName, firstName, lastName, true)

		if err != nil {
			log.Printf("Could not seed faker user: %v", err)
			continue
		}
	}

	fmt.Println("✅ User seeding and mocking completed!")
}
