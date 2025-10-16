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
			// Initial Players
			{ID: uuid.NewString(), Name: "Bima Santoso", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Jakarta Chess Club"},
			{ID: uuid.NewString(), Name: "Putri Anggraini", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Bandung Chess Academy"},
			{ID: uuid.NewString(), Name: "Arif Hidayat", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Surabaya Chess Club"},
			{ID: uuid.NewString(), Name: "Dewi Lestari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Yogyakarta Chess Club"},
			{ID: uuid.NewString(), Name: "Rizky Firmansyah", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Medan Chess Club"},
			{ID: uuid.NewString(), Name: "Siti Rahmawati", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Semarang Chess Academy"},
			{ID: uuid.NewString(), Name: "Eko Prasetyo", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Malang Chess Club"},
			{ID: uuid.NewString(), Name: "Maya Sari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Denpasar Chess Club"},
			{ID: uuid.NewString(), Name: "Joko Susilo", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Solo Chess Academy"},
			{ID: uuid.NewString(), Name: "Dian Permata", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Palembang Chess Club"},
			{ID: uuid.NewString(), Name: "Fajar Nugraha", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Makassar Chess Club"},
			{ID: uuid.NewString(), Name: "Citra Dewi", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Balikpapan Chess Academy"},
			{ID: uuid.NewString(), Name: "Aldo Saputra", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Pontianak Chess Club"},
			{ID: uuid.NewString(), Name: "Rina Melati", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ProgressiveScore: 0, HeadToHeadResults: make(model.HeadToHeadMap), ColorHistory: "", HasBye: false, Club: "Manado Chess Club"},
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
