package tournament

import (
	"encoding/json"
	"sort"
	"time"

	"xchess-desktop/internal/model"
	"xchess-desktop/internal/pkg/utils"
)

// GetPlayers deserializes the PlayersData field into a slice of Player structs.
func GetPlayers(t model.Tournament) ([]model.Player, error) {
	var players []model.Player
	if t.PlayersData == nil {
		return players, nil
	}
	err := json.Unmarshal(t.PlayersData, &players)
	return players, err
}

// SetPlayers serializes a slice of Player structs into the PlayersData field.
func SetPlayers(t *model.Tournament, players []model.Player) error {
	data, err := json.Marshal(players)
	if err != nil {
		return err
	}
	t.PlayersData = data
	return nil
}

// GetRounds deserializes the RoundsData field into a slice of Round structs.
func GetRounds(t model.Tournament) ([]model.Round, error) {
	var rounds []model.Round
	if t.RoundsData == nil {
		return rounds, nil
	}
	err := json.Unmarshal(t.RoundsData, &rounds)
	return rounds, err
}

// SetRounds serializes a slice of Round structs into the RoundsData field.
func SetRounds(t *model.Tournament, rounds []model.Round) error {
	data, err := json.Marshal(rounds)
	if err != nil {
		return err
	}
	t.RoundsData = data
	return nil
}

// PairingEngine abstracts pairing generation so we can adapt different Swiss pairing tools.
type PairingEngine interface {
	GeneratePairings(t *model.Tournament, players []model.Player, roundNumber int) ([]model.Match, error)
}

// SwissToolAdapter is an adapter entry-point for integrating the external "swiss-tool" library.
// TODO: Replace the fallback implementation below with real calls to the swiss-tool package.
type SwissToolAdapter struct{}

// GeneratePairings integrates swisstool for Round 1 and uses model-driven Swiss for later rounds.
func (a SwissToolAdapter) GeneratePairings(t *model.Tournament, players []model.Player, roundNumber int) ([]model.Match, error) {
	// Round 1: use swisstool random pairing directly
	if roundNumber == 1 {
		st := utils.NewTournamentWithConfig(utils.DefaultConfig())
		// Add players using stable order; map utils IDs (1-based) to our players slice index
		for i := range players {
			// Use player ID to avoid duplicate-name constraints internally
			_ = st.AddPlayer(players[i].ID)
		}
		if err := st.StartTournament(); err != nil {
			return nil, err
		}
		pairings := st.GetRound()

		matches := make([]model.Match, 0, len(pairings))
		for i, p := range pairings {
			table := i + 1
			if p.IsBye() {
				aID := players[p.PlayerA()-1].ID
				matches = append(matches, model.Match{
					RoundNumber: roundNumber,
					TableNumber: table,
					PlayerA_ID:  aID,
					PlayerB_ID:  ByePlayerID,
					WhiteID:     aID,
					BlackID:     "",
					Result:      "",
				})
				continue
			}
			aID := players[p.PlayerA()-1].ID
			bID := players[p.PlayerB()-1].ID
			matches = append(matches, model.Match{
				RoundNumber: roundNumber,
				TableNumber: table,
				PlayerA_ID:  aID,
				PlayerB_ID:  bID,
				WhiteID:     aID,
				BlackID:     bID,
				Result:      "",
			})
		}
		return matches, nil
	}

	// Subsequent rounds: Swiss-like pairing using our model state
	// 1) Sort by Score desc, then Buchholz desc (updated via UpdateStandings), tie-break by Rating desc, then Name asc
	ps := make([]model.Player, len(players))
	copy(ps, players)
	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].Score != ps[j].Score {
			return ps[i].Score > ps[j].Score
		}
		if ps[i].Buchholz != ps[j].Buchholz {
			return ps[i].Buchholz > ps[j].Buchholz
		}
		if ps[i].Rating != ps[j].Rating {
			return ps[i].Rating > ps[j].Rating
		}
		return ps[i].Name < ps[j].Name
	})

	// 2) Build quick index: id -> player and helper sets
	index := make(map[string]*model.Player, len(ps))
	for i := range ps {
		index[ps[i].ID] = &ps[i]
	}

	paired := make(map[string]bool, len(ps))
	matches := make([]model.Match, 0, len(ps)/2+1)
	table := 1

	// Helper: check if two players have played before
	havePlayed := func(a, b *model.Player) bool {
		for _, oid := range a.OpponentIDs {
			if oid == b.ID {
				return true
			}
		}
		return false
	}

	// Helper: choose bye candidate (lowest score, no prior bye if possible)
	chooseBye := func(candidates []model.Player) *model.Player {
		// Prefer lowest score and HasBye == false
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Score != candidates[j].Score {
				return candidates[i].Score < candidates[j].Score
			}
			if candidates[i].Buchholz != candidates[j].Buchholz {
				return candidates[i].Buchholz < candidates[j].Buchholz
			}
			return candidates[i].Name < candidates[j].Name
		})
		for i := range candidates {
			if !candidates[i].HasBye {
				return &candidates[i]
			}
		}
		// If all have had bye, pick the absolute lowest
		if len(candidates) > 0 {
			return &candidates[0]
		}
		return nil
	}

	// 3) Pair iteration, prefer same-score opponents without rematches, otherwise closest-score opponents
	for i := 0; i < len(ps); i++ {
		a := &ps[i]
		if paired[a.ID] {
			continue
		}

		// Find best opponent
		var best *model.Player

		// First pass: same-score candidates, no rematch
		for j := i + 1; j < len(ps); j++ {
			b := &ps[j]
			if paired[b.ID] {
				continue
			}
			if a.Score == b.Score && !havePlayed(a, b) {
				best = b
				break
			}
		}
		// Second pass: closest-score candidates, no rematch
		if best == nil {
			for j := i + 1; j < len(ps); j++ {
				b := &ps[j]
				if paired[b.ID] {
					continue
				}
				if !havePlayed(a, b) {
					best = b
					break
				}
			}
		}
		// Last resort: allow rematch
		if best == nil {
			for j := i + 1; j < len(ps); j++ {
				b := &ps[j]
				if paired[b.ID] {
					continue
				}
				best = b
				break
			}
		}

		if best != nil {
			// Simple color assignment with minimal balancing using ColorHistory
			white := a
			black := best
			if len(white.ColorHistory) > 0 && white.ColorHistory[len(white.ColorHistory)-1] == 'W' {
				// If a recently had White, try give Black
				white, black = black, white
			}

			matches = append(matches, model.Match{
				RoundNumber: roundNumber,
				TableNumber: table,
				PlayerA_ID:  a.ID,
				PlayerB_ID:  best.ID,
				WhiteID:     white.ID,
				BlackID:     black.ID,
				Result:      "",
			})
			paired[a.ID] = true
			paired[best.ID] = true
			table++
		}
	}

	// 4) BYE handling if someone remains unpaired
	if len(matches)*2 < len(ps) {
		// Collect unpaired candidates
		var candidates []model.Player
		for i := range ps {
			if !paired[ps[i].ID] {
				candidates = append(candidates, ps[i])
			}
		}
		if bye := chooseBye(candidates); bye != nil {
			matches = append(matches, model.Match{
				RoundNumber: roundNumber,
				TableNumber: table,
				PlayerA_ID:  bye.ID,
				PlayerB_ID:  ByePlayerID,
				WhiteID:     bye.ID,
				BlackID:     "",
				Result:      "",
			})
		}
	}

	return matches, nil
}

const ByePlayerID = "BYE"

// InitializeTournament sets minimal fields and attaches players.
// Title is required; players will be serialized into PlayersData.
// PairingSystem defaults to "SWISS"; ByeScore defaults to 1.0 if unset.
func InitializeTournament(t *model.Tournament, title string, players []model.Player) error {
	t.Title = title
	t.Description = "Swiss-system tournament"
	t.Status = "ACTIVE"
	t.StartTime = time.Now()
	t.CurrentRound = 0
	t.TotalPlayers = len(players)
	if t.PairingSystem == "" {
		t.PairingSystem = "SWISS"
	}
	if t.ByeScore == 0 {
		t.ByeScore = 1.0
	}

	// Persist players
	if err := t.SetPlayers(players); err != nil {
		return err
	}

	// Initialize empty rounds
	if err := t.SetRounds([]model.Round{}); err != nil {
		return err
	}

	return nil
}

// RecordMatchResult updates the specified match result and player standings.
// result must be one of: "A_WIN", "B_WIN", "DRAW", "BYE_A".
func RecordMatchResult(t *model.Tournament, roundNumber int, tableNumber int, result string) error {
	rounds, err := t.GetRounds()
	if err != nil {
		return err
	}
	players, err := t.GetPlayers()
	if err != nil {
		return err
	}

	// Build quick index for players
	playerIndex := make(map[string]*model.Player, len(players))
	for i := range players {
		p := &players[i]
		playerIndex[p.ID] = p
	}

	// Locate the target match
	var match *model.Match
	for r := range rounds {
		if rounds[r].RoundNumber != roundNumber {
			continue
		}
		for m := range rounds[r].Matches {
			if rounds[r].Matches[m].TableNumber == tableNumber {
				match = &rounds[r].Matches[m]
				break
			}
		}
		if match != nil {
			break
		}
	}
	if match == nil {
		return nil // No-op if not found; you may change to error if preferred
	}

	// Update match scores based on result
	switch result {
	case "A_WIN":
		match.Result = "A_WIN"
		match.ScoreA = 1.0
		match.ScoreB = 0.0
	case "B_WIN":
		match.Result = "B_WIN"
		match.ScoreA = 0.0
		match.ScoreB = 1.0
	case "DRAW":
		match.Result = "DRAW"
		match.ScoreA = 0.5
		match.ScoreB = 0.5
	case "BYE_A":
		match.Result = "BYE_A"
		if t.ByeScore == 0 {
			t.ByeScore = 1.0
		}
		match.ScoreA = t.ByeScore
		match.ScoreB = 0.0
	default:
		// Unknown result; no changes
		return nil
	}

	// Apply player updates
	aID := match.PlayerA_ID
	bID := match.PlayerB_ID

	// Update scores and opponent lists
	if a, ok := playerIndex[aID]; ok {
		a.Score += match.ScoreA
		if bID != ByePlayerID {
			ensureOpponent(a, bID)
			// Update color history: "W" if a is White, "B" if Black
			if match.WhiteID == a.ID {
				a.ColorHistory += "W"
			} else if match.BlackID == a.ID {
				a.ColorHistory += "B"
			}
		} else {
			a.HasBye = true
			// You may add a marker for bye in color history if desired, e.g., "-"
		}
	}
	if bID != ByePlayerID {
		if b, ok := playerIndex[bID]; ok {
			b.Score += match.ScoreB
			ensureOpponent(b, aID)
			if match.WhiteID == b.ID {
				b.ColorHistory += "W"
			} else if match.BlackID == b.ID {
				b.ColorHistory += "B"
			}
		}
	}

	// Persist updated rounds and players
	if err := t.SetRounds(rounds); err != nil {
		return err
	}
	if err := t.SetPlayers(players); err != nil {
		return err
	}

	// Recompute Buchholz after this result
	UpdateStandings(t)

	return nil
}

func ensureOpponent(p *model.Player, oid string) {
	for _, id := range p.OpponentIDs {
		if id == oid {
			return
		}
	}
	p.OpponentIDs = append(p.OpponentIDs, oid)
}

// UpdateStandings recomputes Buchholz for all players based on OpponentIDs and current scores.
func UpdateStandings(t *model.Tournament) error {
	players, err := t.GetPlayers()
	if err != nil {
		return err
	}

	// Build score index
	scoreIndex := make(map[string]float64, len(players))
	for _, p := range players {
		scoreIndex[p.ID] = p.Score
	}

	for i := range players {
		sum := 0.0
		for _, oid := range players[i].OpponentIDs {
			// Skip bye opponent for Buchholz
			if oid == ByePlayerID {
				continue
			}
			sum += scoreIndex[oid]
		}
		players[i].Buchholz = sum
	}

	return t.SetPlayers(players)
}

// AdvanceToNextRound runs the pairing engine for the next round and persists the round.
// It updates CurrentRound and TotalPlayers on the tournament.
func AdvanceToNextRound(t *model.Tournament, engine PairingEngine) error {
	players, err := t.GetPlayers()
	if err != nil {
		return err
	}

	nextRoundNumber := t.CurrentRound + 1

	// Pass the tournament to the pairing engine for context
	matches, err := engine.GeneratePairings(t, players, nextRoundNumber)
	if err != nil {
		return err
	}

	rounds, err := t.GetRounds()
	if err != nil {
		return err
	}

	newRound := model.Round{
		RoundNumber: nextRoundNumber,
		Matches:     matches,
		IsComplete:  false,
	}
	rounds = append(rounds, newRound)

	if err := t.SetRounds(rounds); err != nil {
		return err
	}

	t.CurrentRound = nextRoundNumber
	t.TotalPlayers = len(players)

	return nil
}
