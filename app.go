package main

import (
	"context"
	"log"
	"xchess-desktop/internal/auth"
	"xchess-desktop/internal/database"

	"xchess-desktop/internal/model"
	"xchess-desktop/internal/tournament"

	"github.com/google/uuid"
)

// App struct is the main application structure for Wails
type App struct {
	ctx               context.Context
	currentTournament *model.Tournament
	engine            tournament.PairingEngine
	db                *database.DB
	authSvc           *auth.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize default pairing engine (Swiss)
	a.engine = tournament.SwissToolAdapter{}

	// Initialize database and services
	dbPath, err := database.GetDBPath()
	if err != nil {
		log.Printf("failed to get DB path: %v", err)
	} else {
		a.db, err = database.New(dbPath)
		if err != nil {
			log.Printf("failed to open DB: %v", err)
		} else {
			if err = a.db.RunMigrations(); err != nil {
				log.Printf("failed to run migrations: %v", err)
			}
			a.authSvc, err = auth.New(a.db)
			if err != nil {
				log.Printf("failed to init auth service: %v", err)
			}
		}
	}

	log.Printf("App startup complete")
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
}

// CheckAdminCredentials delegates to the auth service
func (a *App) CheckAdminCredentials(username, password string) (bool, error) {
	if a.authSvc == nil {
		log.Printf("App.CheckAdminCredentials: authSvc is nil; login failed for user=%q", username)
		return false, nil
	}
	log.Printf("App.CheckAdminCredentials: login attempt user=%q", username)
	ok, err := a.authSvc.CheckCredentials(username, password)
	if err != nil {
		log.Printf("App.CheckAdminCredentials: error for user=%q: %v", username, err)
		return false, err
	}
	log.Printf("App.CheckAdminCredentials: result for user=%q: %t", username, ok)
	return ok, nil
}

// Initialize a new tournament with a title and player names.
// Returns true if initialization succeeded.
func (a *App) InitTournament(title string, playerNames []string) (bool, error) {
	players := make([]model.Player, 0, len(playerNames))
	for _, name := range playerNames {
		players = append(players, model.Player{
			ID:           uuid.NewString(),
			Name:         name,
			Score:        0,
			OpponentIDs:  []string{},
			Buchholz:     0,
			ColorHistory: "",
			HasBye:       false,
		})
	}

	t := &model.Tournament{
		ByeScore:      1.0,
		PairingSystem: "SWISS",
	}
	if err := tournament.InitializeTournament(t, title, players); err != nil {
		return false, err
	}
	a.currentTournament = t
	return true, nil
}

// Advance to the next round and generate pairings.
// Returns true if the round was generated.
func (a *App) NextRound() (bool, error) {
	if a.currentTournament == nil {
		return false, nil
	}
	if err := tournament.AdvanceToNextRound(a.currentTournament, a.engine); err != nil {
		return false, err
	}
	return true, nil
}

// Get the current round matches.
func (a *App) GetCurrentRound() (model.Round, error) {
	var empty model.Round
	if a.currentTournament == nil {
		return empty, nil
	}
	rounds, err := a.currentTournament.GetRounds()
	if err != nil {
		return empty, err
	}
	cr := a.currentTournament.CurrentRound
	for _, r := range rounds {
		if r.RoundNumber == cr {
			return r, nil
		}
	}
	return empty, nil
}

// Record a result for a given table in the current round.
// result must be one of: "A_WIN", "B_WIN", "DRAW", "BYE_A".
func (a *App) RecordResult(tableNumber int, result string) (bool, error) {
	if a.currentTournament == nil {
		return false, nil
	}
	cr := a.currentTournament.CurrentRound
	if err := tournament.RecordMatchResult(a.currentTournament, cr, tableNumber, result); err != nil {
		return false, err
	}
	return true, nil
}

// Get the current players (including scores and buchholz).
func (a *App) GetPlayers() ([]model.Player, error) {
	if a.currentTournament == nil {
		return []model.Player{}, nil
	}
	return a.currentTournament.GetPlayers()
}

// Optionally expose basic tournament info for the frontend.
func (a *App) GetTournamentInfo() (model.Tournament, error) {
	if a.currentTournament == nil {
		return model.Tournament{}, nil
	}
	return *a.currentTournament, nil
}

// ListPlayers returns all players (peserta) from the database for selection in the frontend.
func (a *App) ListPlayers() ([]model.Player, error) {
	if a.db == nil {
		return []model.Player{}, nil
	}
	var players []model.Player
	if err := a.db.Find(&players).Error; err != nil {
		return []model.Player{}, err
	}
	return players, nil
}

// Initialize a new tournament using selected existing player IDs.
// No player creation; we load players from the DB and initialize the tournament.
func (a *App) InitTournamentWithPlayerIDs(title string, playerIDs []string) (bool, error) {
	if a.db == nil {
		return false, nil
	}
	var players []model.Player
	if err := a.db.Where("id IN ?", playerIDs).Find(&players).Error; err != nil {
		return false, err
	}
	if len(players) == 0 {
		return false, nil
	}

	t := &model.Tournament{
		ByeScore:      1.0,
		PairingSystem: "SWISS",
	}
	if err := tournament.InitializeTournament(t, title, players); err != nil {
		return false, err
	}
	a.currentTournament = t
	return true, nil
}
