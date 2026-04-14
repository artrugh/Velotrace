//go:build devseed

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/velotrace/bikes-api/internal/domain"
	"velotrace.local/logger"
)

func main() {
	logger.Init("bikes-seeder")
	l := logger.L

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/identity?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		l.Error("unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	l.Info("🌱 starting bike seeding...")

	var userIDs []uuid.UUID
	rows, err := pool.Query(ctx, "SELECT id FROM users")
	if err != nil {
		l.Error("no users found. seed identity-api first!", "err", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			l.Error("error scanning user id", "err", err)
			continue
		}
		userIDs = append(userIDs, id)
	}

	if len(userIDs) == 0 {
		l.Error("no users found in database")
		os.Exit(1)
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
			l.Warn("could not seed bike", "err", err)
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
				l.Warn("could not seed bike image", "err", err)
			}
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO ownership_records (id, bike_id, owner_id, is_active)
			VALUES ($1, $2, $3, $4)
		`, uuid.New(), bikeID, ownerID, true)
		if err != nil {
			l.Warn("could not seed ownership record", "err", err)
		}

		l.Info("created bike", "make_model", makeModel, "owner", ownerID)
	}

	l.Info("✅ bike seeding completed!")
}
