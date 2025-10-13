package database

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"xchess-desktop/internal/model"

	"gorm.io/gorm"
)

// RunMigrations performs database migrations for all models using GORM
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Enable foreign key constraints in SQLite
	db.Exec("PRAGMA foreign_keys = ON;")

	// Use GORM's AutoMigrate to handle all migrations
	err := db.AutoMigrate(
		&model.Administrator{},
		&model.Player{},
		&model.Match{},
		&model.Round{},
		&model.Tournament{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate models: %v", err)
	}
	log.Println("GORM AutoMigrate completed for all models.")

	// Seed initial data
	if err := SeedInitialData(db); err != nil {
		return fmt.Errorf("failed to seed initial data: %v", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// SeedInitialData seeds the database with minimal, essential data
func SeedInitialData(db *gorm.DB) error {
	log.Println("Seeding initial data...")

	// Ensure a single administrator exists; create if missing
	var admin model.Administrator
	err := db.Where("username = ?", "admin").First(&admin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			hashedPassword, herr := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
			if herr != nil {
				return fmt.Errorf("failed to hash initial administrator password: %v", herr)
			}
			admin = model.Administrator{
				ID:       uuid.New(),
				Username: "admin",
				Password: string(hashedPassword),
				Role:     model.Admin,
			}
			if createErr := db.Create(&admin).Error; createErr != nil {
				return fmt.Errorf("failed to create initial administrator: %v", err)
			}
			log.Println("Initial administrator seeded successfully")
		} else {
			return fmt.Errorf("failed to query administrator: %v", err)
		}
	}

	// Seed initial players only if none exist
	var count int64
	if err := db.Model(&model.Player{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count players: %v", err)
	}
	if count == 0 {
		initialPlayers := []model.Player{
			// Initial 8 Players
			{ID: uuid.NewString(), Name: "Michael Chen", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1600},
			{ID: uuid.NewString(), Name: "Emily Johnson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1650},
			{ID: uuid.NewString(), Name: "David Smith", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1700},
			{ID: uuid.NewString(), Name: "Sarah Kim", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1580},
			{ID: uuid.NewString(), Name: "James Wilson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1750},
			{ID: uuid.NewString(), Name: "Jessica Brown", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1620},
			{ID: uuid.NewString(), Name: "Robert Garcia", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1680},
			{ID: uuid.NewString(), Name: "Olivia Martin", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1550},
			{ID: uuid.NewString(), Name: "William Lee", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1800},
			{ID: uuid.NewString(), Name: "Ava Miller", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1520},
			{ID: uuid.NewString(), Name: "Noah Davis", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1720},
			{ID: uuid.NewString(), Name: "Sophia Rodriguez", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1640},
			{ID: uuid.NewString(), Name: "Liam Jones", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1780},
			{ID: uuid.NewString(), Name: "Isabella Martinez", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1510},
			{ID: uuid.NewString(), Name: "Ethan Hernandez", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1690},
			{ID: uuid.NewString(), Name: "Mia Lopez", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1610},
			{ID: uuid.NewString(), Name: "Alexander Gonzalez", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1850},
			{ID: uuid.NewString(), Name: "Charlotte Pérez", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1540},
			{ID: uuid.NewString(), Name: "Jacob Torres", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1760},
			{ID: uuid.NewString(), Name: "Amelia Flores", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1490},
			{ID: uuid.NewString(), Name: "Daniel Rivera", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1660},
			{ID: uuid.NewString(), Name: "Evelyn Cox", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1570},
			{ID: uuid.NewString(), Name: "Matthew Peterson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1740},
			{ID: uuid.NewString(), Name: "Harper Kelly", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1480},
			{ID: uuid.NewString(), Name: "Andrew Reed", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1820},
			{ID: uuid.NewString(), Name: "Ella Rogers", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1590},
			{ID: uuid.NewString(), Name: "Joseph Cooper", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1710},
			{ID: uuid.NewString(), Name: "Scarlett Bailey", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1470},
			{ID: uuid.NewString(), Name: "Samuel Bell", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1630},
			{ID: uuid.NewString(), Name: "Victoria Murphy", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1880},
			{ID: uuid.NewString(), Name: "Henry Cook", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1500},
			{ID: uuid.NewString(), Name: "Zoe Howard", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1730},
			{ID: uuid.NewString(), Name: "Gabriel Watson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1460},
			{ID: uuid.NewString(), Name: "Avery Parker", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1670},
			{ID: uuid.NewString(), Name: "Jack Nelson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1530},
			{ID: uuid.NewString(), Name: "Layla Carter", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1790},
			{ID: uuid.NewString(), Name: "Owen Evans", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1450},
			{ID: uuid.NewString(), Name: "Grace Hall", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1830},
			{ID: uuid.NewString(), Name: "Luke Baker", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1560},
			{ID: uuid.NewString(), Name: "Chloe Green", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1770},
			{ID: uuid.NewString(), Name: "Ryan Morris", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1440},
			{ID: uuid.NewString(), Name: "Madison King", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1810},
			{ID: uuid.NewString(), Name: "Carter Scott", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1605},
			{ID: uuid.NewString(), Name: "Lily Adams", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1430},
			{ID: uuid.NewString(), Name: "Christopher Wright", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1705},
			{ID: uuid.NewString(), Name: "Eleanor Hill", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1645},
			{ID: uuid.NewString(), Name: "Isaac Wood", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1860},
			{ID: uuid.NewString(), Name: "Hazel Allen", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1420},
			{ID: uuid.NewString(), Name: "Joshua Thompson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1745},
			{ID: uuid.NewString(), Name: "Penelope Clark", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1685},
		}
		for _, p := range initialPlayers {
			if err := db.Create(&p).Error; err != nil {
				return fmt.Errorf("failed to seed player %s: %v", p.Name, err)
			}
		}
		log.Println("Initial players seeded successfully")
	}

	return nil
}
