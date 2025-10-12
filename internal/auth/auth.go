package auth

import (
	"fmt"

	"xchess-desktop/internal/database"
	"xchess-desktop/internal/model"

	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service manages authentication operations
type Service struct {
	db *database.DB
}

// New creates a new authentication service
func New(db *database.DB) (*Service, error) {
	service := &Service{
		db: db,
	}

	return service, nil
}

// CheckCredentials checks if the provided username and password are valid
func (s *Service) CheckCredentials(username, password string) (bool, error) {
	log.Printf("auth: CheckCredentials called: username=%q (password length=%d)", username, len(password))

	var admin model.Administrator

	// Find admin by username
	result := s.db.Where("username = ?", username).First(&admin)

	// Check if admin exists
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("auth: user not found: %q", username)
			return false, nil // User not found
		}
		log.Printf("auth: database query error for user=%q: %v", username, result.Error)
		return false, fmt.Errorf("database query error: %w", result.Error)
	}

	// Log stored password characteristics (do not log the password itself)
	stored := admin.Password
	hashed := strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$")
	log.Printf("auth: user=%q found; stored password len=%d; hashed=%t", username, len(stored), hashed)

	// Also log bcrypt cost of stored hash, if parsable
	if cost, cerr := bcrypt.Cost([]byte(stored)); cerr == nil {
		log.Printf("auth: stored hash cost for user=%q: %d", username, cost)
	} else {
		log.Printf("auth: unable to parse bcrypt cost for user=%q: %v", username, cerr)
	}

	// Compare password
	err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(password))
	if err != nil {
		log.Printf("auth: bcrypt compare failed for user=%q: %v", username, err)
		return false, nil // Password does not match
	}

	log.Printf("auth: bcrypt compare succeeded for user=%q", username)
	return true, nil // Credentials are valid
}
