//go:build devseed

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/models"
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

	fmt.Println("🌱 Seeding Bikes...")

	// 1. Get existing user IDs to assign owners
	var userIDs []uuid.UUID
	err = db.Table("users").Select("id").Find(&userIDs).Error
	if err != nil || len(userIDs) == 0 {
		log.Fatalf("No users found to assign bikes. Seed identity-api first! Error: %v", err)
	}

	bikeMakes := []string{"Specialized", "Trek", "Giant", "Cannondale", "Canyon", "Santa Cruz", "Scott", "Bianchi"}
	statuses := []models.BikeStatus{models.StatusRegistered, models.StatusForSale, models.StatusStolen}

	for i := 0; i < 20; i++ {
		ownerID := userIDs[rand.Intn(len(userIDs))]
		makeModel := fmt.Sprintf("%s %s", bikeMakes[rand.Intn(len(bikeMakes))], faker.Word())

		bike := models.Bike{
			ID:             uuid.New(),
			MakeModel:      makeModel,
			Year:           2015 + rand.Intn(11),
			Price:          float64(500 + rand.Intn(5000)),
			LocationCity:   faker.Word(),
			CurrentOwnerID: ownerID,
			SerialNumber:   faker.UUIDDigit(),
			Description:    faker.Sentence(),
			Status:         statuses[rand.Intn(len(statuses))],
		}

		if err := db.Create(&bike).Error; err != nil {
			log.Printf("Could not seed bike: %v", err)
			continue
		}

		// 2. Seed some images
		for j := 0; j < 2; j++ {
			img := models.BikeImage{
				ID:        uuid.New(),
				BikeID:    bike.ID,
				URL:       fmt.Sprintf("https://picsum.photos/seed/%s/800/600", uuid.New().String()),
				IsPrimary: j == 0,
			}
			db.Create(&img)
		}

		// 3. Seed ownership record
		record := models.OwnershipRecord{
			ID:       uuid.New(),
			BikeID:   bike.ID,
			OwnerID:  ownerID,
			IsActive: true,
		}
		db.Create(&record)

		fmt.Printf("Created bike: %s for owner %s\n", bike.MakeModel, ownerID)
	}

	fmt.Println("✅ Bike seeding completed!")
}
