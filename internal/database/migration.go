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
			{ID: uuid.NewString(), Name: "Michael Chen", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1600},
			{ID: uuid.NewString(), Name: "Emily Johnson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1650},
			{ID: uuid.NewString(), Name: "David Smith", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1700},
			{ID: uuid.NewString(), Name: "Sarah Kim", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1580},
			{ID: uuid.NewString(), Name: "James Wilson", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1750},
			{ID: uuid.NewString(), Name: "Jessica Brown", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1620},
			{ID: uuid.NewString(), Name: "Robert Garcia", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1680},
			{ID: uuid.NewString(), Name: "Olivia Martin", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1550},
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
