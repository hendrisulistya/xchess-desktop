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
			{ID: uuid.NewString(), Name: "Bima Santoso", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1600},
			{ID: uuid.NewString(), Name: "Putri Anggraini", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1650},
			{ID: uuid.NewString(), Name: "Arif Hidayat", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1700},
			{ID: uuid.NewString(), Name: "Dewi Lestari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1580},
			{ID: uuid.NewString(), Name: "Rizky Firmansyah", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1750},
			{ID: uuid.NewString(), Name: "Siti Rahmawati", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1620},
			{ID: uuid.NewString(), Name: "Eko Prasetyo", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1680},
			{ID: uuid.NewString(), Name: "Maya Sari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1550},
			{ID: uuid.NewString(), Name: "Joko Susilo", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1800},
			{ID: uuid.NewString(), Name: "Dian Permata", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1520},
			{ID: uuid.NewString(), Name: "Fajar Nugraha", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1720},
			{ID: uuid.NewString(), Name: "Citra Dewi", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1640},
			{ID: uuid.NewString(), Name: "Aldo Saputra", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1780},
			{ID: uuid.NewString(), Name: "Rina Melati", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1510},
			{ID: uuid.NewString(), Name: "Dani Iskandar", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1690},
			{ID: uuid.NewString(), Name: "Indah Jaya", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1610},
			{ID: uuid.NewString(), Name: "Gilang Ramadhan", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1850},
			{ID: uuid.NewString(), Name: "Nadia Chandra", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1540},
			{ID: uuid.NewString(), Name: "Bayu Wijaya", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1760},
			{ID: uuid.NewString(), Name: "Lina Fauzi", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1490},
			{ID: uuid.NewString(), Name: "Taufik Rahman", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1660},
			{ID: uuid.NewString(), Name: "Ayu Wulandari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1570},
			{ID: uuid.NewString(), Name: "Wayan Putra", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1740},
			{ID: uuid.NewString(), Name: "Kirana Dewi", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1480},
			{ID: uuid.NewString(), Name: "Dwi Nugroho", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1820},
			{ID: uuid.NewString(), Name: "Gita Sanjaya", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1590},
			{ID: uuid.NewString(), Name: "Aditya Pratama", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1710},
			{ID: uuid.NewString(), Name: "Sekar Tani", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1470},
			{ID: uuid.NewString(), Name: "Arianto Budiman", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1630},
			{ID: uuid.NewString(), Name: "Kartika Sari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1880},
			{ID: uuid.NewString(), Name: "Cahyo Utomo", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1500},
			{ID: uuid.NewString(), Name: "Melati Kusuma", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1730},
			{ID: uuid.NewString(), Name: "Akbar Maulana", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1460},
			{ID: uuid.NewString(), Name: "Rani Suryani", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1670},
			{ID: uuid.NewString(), Name: "Hendra Gunawan", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1530},
			{ID: uuid.NewString(), Name: "Sinta Wijaya", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1790},
			{ID: uuid.NewString(), Name: "Irfan Hakim", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1450},
			{ID: uuid.NewString(), Name: "Larasati Putri", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1830},
			{ID: uuid.NewString(), Name: "Bagus Setiawan", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1560},
			{ID: uuid.NewString(), Name: "Tania Hapsari", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1770},
			{ID: uuid.NewString(), Name: "Denny Maulana", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1440},
			{ID: uuid.NewString(), Name: "Annisa Dewi", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1810},
			{ID: uuid.NewString(), Name: "Dimas Cahyono", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1605},
			{ID: uuid.NewString(), Name: "Mila Adiwijaya", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1430},
			{ID: uuid.NewString(), Name: "Yoga Pradana", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1705},
			{ID: uuid.NewString(), Name: "Kania Nurmala", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1645},
			{ID: uuid.NewString(), Name: "Andre Wijaya", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1860},
			{ID: uuid.NewString(), Name: "Intan Pratiwi", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1420},
			{ID: uuid.NewString(), Name: "Eka Setiawan", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1745},
			{ID: uuid.NewString(), Name: "Ratna Kumala", Score: 0, OpponentIDs: []string{}, Buchholz: 0, ColorHistory: "", HasBye: false, Rating: 1685},
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
