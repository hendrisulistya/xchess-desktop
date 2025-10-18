// internal/database/database.go
package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB manages database operations with GORM
type DB struct {
	*gorm.DB
	dbPath string
}

// GetDBPath determines the appropriate path for the SQLite database file
func GetDBPath() (string, error) {
	// Get the user's configuration directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	// Create a subdirectory for your application's data
	appDataDir := filepath.Join(configDir, "xchess-data")

	// Ensure the directory exists with secure permissions
	err = os.MkdirAll(appDataDir, 0700) // 0700 = owner-only access
	if err != nil {
		return "", fmt.Errorf("failed to create application data directory: %w", err)
	}

	// Construct the full path to the database file
	dbPath := filepath.Join(appDataDir, "app.db")

	// If the file already exists, ensure it has secure permissions
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.Chmod(dbPath, 0600); err != nil { // 0600 = owner read/write only
			log.Printf("Warning: Could not set secure permissions on database file: %v", err)
		}
	}

	return dbPath, nil
}

// New creates a standard unencrypted database
func New(dbPath string) (*DB, error) {
	log.Printf("Initializing database connection at: %s", dbPath)

	// Configure GORM with better settings for Windows
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open SQLite database with additional pragmas for Windows compatibility
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=FULL&_cache_size=1000&_foreign_keys=on", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Get underlying SQL DB for additional configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database: %w", err)
	}

	// Configure connection pool for better performance
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection
	sqlDB.SetMaxIdleConns(1)

	// Set secure pragmas
	db.Exec("PRAGMA secure_delete = ON;")

	return &DB{DB: db, dbPath: dbPath}, nil
}

// RunMigrations now calls the RunMigrations function from the migrations package
func (db *DB) RunMigrations() error {
	return RunMigrations(db.DB)
}

// Close closes the database connection
func (db *DB) Close() error {
	log.Println("Closing database connection...")
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
