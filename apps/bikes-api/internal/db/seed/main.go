//go:build devseed

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/velotrace/bikes-api/internal/domain"
)

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

	fmt.Println("🌱 Seeding Bikes...")

	var userIDs []uuid.UUID
	rows, err := pool.Query(ctx, "SELECT id FROM users")
	if err != nil {
		log.Fatalf("No users found to assign bikes. Seed identity-api first! Error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("Error scanning user ID: %v", err)
		}
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		log.Fatalf("No users found in database.")
	}

	bikeMakes := []string{"Specialized", "Trek", "Giant", "Cannondale", "Canyon", "Santa Cruz", "Scott", "Bianchi"}
	statuses := []domain.BikeStatus{domain.StatusRegistered, domain.StatusForSale, domain.StatusStolen}

	for i := 0; i < 20; i++ {
		ownerID := userIDs[rand.Intn(len(userIDs))]
		makeModel := fmt.Sprintf("%s %s", bikeMakes[rand.Intn(len(bikeMakes))], faker.Word())

		bikeID := uuid.New()
		status := statuses[rand.Intn(len(statuses))]
		year := 2015 + rand.Intn(11)
		price := float64(500 + rand.Intn(5000))
		location := faker.Word()
		serial := faker.UUIDDigit()
		desc := faker.Sentence()

		_, err := pool.Exec(ctx, `
			INSERT INTO bikes (id, make_model, year, price, location_city, current_owner_id, serial_number, description, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, bikeID, makeModel, year, price, location, ownerID, serial, desc, status)

		if err != nil {
			log.Printf("Could not seed bike: %v", err)
			continue
		}

		for j := 0; j < 2; j++ {
			imgID := uuid.New()
			objKey := fmt.Sprintf("bikes/seed/%s.jpg", uuid.New().String())
			isPrimary := j == 0
			_, err := pool.Exec(ctx, `
				INSERT INTO bike_images (id, bike_id, object_key, is_primary)
				VALUES ($1, $2, $3, $4)
			`, imgID, bikeID, objKey, isPrimary)
			if err != nil {
				log.Printf("Could not seed bike image: %v", err)
			}
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO ownership_records (id, bike_id, owner_id, is_active)
			VALUES ($1, $2, $3, $4)
		`, uuid.New(), bikeID, ownerID, true)
		if err != nil {
			log.Printf("Could not seed ownership record: %v", err)
		}

		fmt.Printf("Created bike: %s for owner %s\n", makeModel, ownerID)
	}

	fmt.Println("✅ Bike seeding completed!")
}
