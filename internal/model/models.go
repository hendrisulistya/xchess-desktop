package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ===== ENUMS =====

// Role represents administrator access levels
type Role string

const (
	Sudo  Role = "SUDO"
	Admin Role = "ADMIN"
)

// ===== MODELS =====

// Administrator represents a system administrator with access controls
type Administrator struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	Username  string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	Role      Role      `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Player represents a single participant in the tournament.
type Player struct {
	ID           string   `json:"id" gorm:"primaryKey"` // Unique short ID or player handle
	Name         string   `json:"name"`
	Score        float64  `json:"score"`                         // Current total points (e.g., 1.0 for Win, 0.5 for Draw)
	OpponentIDs  []string `json:"opponent_ids" gorm:"type:json"` // List of IDs of players already faced (Crucial for Swiss Pairing)
	Buchholz     float64  `json:"buchholz"`                      // Primary Tie-breaker: Sum of opponents' scores
	ColorHistory string   `json:"color_history"`                 // E.g., "WBW" (White, Black, White) to track color imbalance
	HasBye       bool     `json:"has_bye"`                       // True if the player has received a bye
	Rating       int      `json:"rating,omitempty"`              // Initial seeding (optional)
}

// Match represents the outcome of a single game between two players.
type Match struct {
	MatchID     uuid.UUID `json:"match_id" gorm:"type:uuid"`
	RoundNumber int       `json:"round_number"`
	TableNumber int       `json:"table_number"` // New field: The physical table where the match is played
	PlayerA_ID  string    `json:"player_a_id"`
	PlayerB_ID  string    `json:"player_b_id"` // Set to "BYE" if a bye is assigned

	// Added to support color balancing and tracking
	WhiteID string `json:"white_id"`
	BlackID string `json:"black_id"`

	Result string  `json:"result"`  // E.g., "A_WIN", "B_WIN", "DRAW", "BYE_A"
	ScoreA float64 `json:"score_a"` // Points awarded to Player A
	ScoreB float64 `json:"score_b"` // Points awarded to Player B
}

// Round encapsulates all matches played in a single step of the tournament.
type Round struct {
	RoundNumber int     `json:"round_number"`
	Matches     []Match `json:"matches" gorm:"type:json"`
	IsComplete  bool    `json:"is_complete"`
}

// Tournament holds the overall state and history of a Swiss-system event.
type Tournament struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null"` // "SETUP", "ACTIVE", "COMPLETE"

	// Core data for Swiss logic (stored as JSON in the database for single record management)
	PlayersData json.RawMessage `json:"players_data" gorm:"column:players;type:json"`
	RoundsData  json.RawMessage `json:"rounds_data" gorm:"column:rounds;type:json"`
	EventsData  json.RawMessage `json:"events_data" gorm:"column:events;type:json"`

	// Summary/Metadata
	CurrentRound int        `json:"current_round" gorm:"not null"`
	TotalPlayers int        `json:"total_players" gorm:"not null"`
	StartTime    time.Time  `json:"start_time" gorm:"not null"`
	EndTime      *time.Time `json:"end_time"` // Nullable: only set when tournament is complete

	// Pairing configuration
	RoundsTotal   int     `json:"rounds_total,omitempty"`
	ByeScore      float64 `json:"bye_score,omitempty"`
	PairingSystem string  `json:"pairing_system,omitempty"` // e.g., "SWISS"

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Event represents a tournament event for audit trail and detailed reporting
type Event struct {
	EventID     uuid.UUID       `json:"event_id"`
	Type        string          `json:"type"`        // e.g., "MATCH_RESULT_RECORDED", "ROUND_STARTED"
	Timestamp   time.Time       `json:"timestamp"`
	RoundNumber int             `json:"round_number"`
	TableNumber int             `json:"table_number,omitempty"`
	Details     json.RawMessage `json:"details,omitempty"` // JSON payload with event-specific data
}
