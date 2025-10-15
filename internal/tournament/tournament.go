/*
Maintainers note:
This file implements Swiss tournament rules and gameplay logic.
Refer to the specification at internal/tournament/tournament.md for the current rules, pairing, scoring, tie-breaks, and lifecycle.
Update implementations here to match the specification as it evolves.
*/
package tournament

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"xchess-desktop/internal/model"
	"xchess-desktop/internal/pkg/utils"

	"github.com/google/uuid"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
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

	// Subsequent rounds: Swiss-like pairing with hard constraints (no rematches, max score diff 1.0)
	ps := make([]model.Player, len(players))
	copy(ps, players)
	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].Score != ps[j].Score {
			return ps[i].Score > ps[j].Score
		}
		if ps[i].Buchholz != ps[j].Buchholz {
			return ps[i].Buchholz > ps[j].Buchholz
		}
		return ps[i].Name < ps[j].Name
	})

	// Compute last table numbers from previous round for table proximity
	lastTable := make(map[string]int, len(ps))
	if roundNumber > 1 {
		if rounds, err := t.GetRounds(); err == nil {
			for _, r := range rounds {
				if r.RoundNumber == roundNumber-1 {
					for _, m := range r.Matches {
						lastTable[m.PlayerA_ID] = m.TableNumber
						if m.PlayerB_ID != ByePlayerID {
							lastTable[m.PlayerB_ID] = m.TableNumber
						}
					}
					break
				}
			}
		}
	}

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

	// Backtracking pairing under constraints
	used := make(map[string]bool, len(ps))
	matches := make([]model.Match, 0, len(ps)/2+1)
	table := 1
	allowBye := len(ps)%2 == 1
	byeAssigned := false

	abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}
	intAbs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	var backtrack func() bool
	backtrack = func() bool {
		// Find first unpaired player
		var a *model.Player
		for i := range ps {
			if !used[ps[i].ID] {
				a = &ps[i]
				break
			}
		}
		// All paired
		if a == nil {
			return true
		}

		// Build candidate list: not used, no rematch, within score diff <= 1.0
		type cand struct {
			j         int
			scoreDiff float64
			tableProx int
		}
		var cands []cand
		for j := range ps {
			if ps[j].ID == a.ID || used[ps[j].ID] {
				continue
			}
			if havePlayed(a, &ps[j]) {
				continue
			}
			diff := abs(a.Score - ps[j].Score)
			if diff > 1.0 {
				continue
			}
			// Prefer pairing with closest previous tables (secondary priority)
			aTable := lastTable[a.ID]
			bTable := lastTable[ps[j].ID]
			prox := 1 << 30
			if aTable > 0 && bTable > 0 {
				prox = intAbs(aTable - bTable)
			}
			cands = append(cands, cand{j: j, scoreDiff: diff, tableProx: prox})
		}
		// Prefer same-score (diff=0), then closest previous tables
		sort.SliceStable(cands, func(i, j int) bool {
			if cands[i].scoreDiff != cands[j].scoreDiff {
				return cands[i].scoreDiff < cands[j].scoreDiff
			}
			return cands[i].tableProx < cands[j].tableProx
		})

		for _, c := range cands {
			b := &ps[c.j]
			white := a
			black := b
			if len(white.ColorHistory) > 0 && white.ColorHistory[len(white.ColorHistory)-1] == 'W' {
				white, black = black, white
			}

			used[a.ID] = true
			used[b.ID] = true
			matches = append(matches, model.Match{
				RoundNumber: roundNumber,
				TableNumber: table,
				PlayerA_ID:  a.ID,
				PlayerB_ID:  b.ID,
				WhiteID:     white.ID,
				BlackID:     black.ID,
				Result:      "",
			})
			table++

			if backtrack() {
				return true
			}

			// Undo
			table--
			matches = matches[:len(matches)-1]
			used[b.ID] = false
			used[a.ID] = false
		}

		// If no candidate found, try assigning BYE only if allowed and not yet assigned.
		// Assign BYE to the actual low-score candidate among remaining unpaired players.
		if allowBye && !byeAssigned {
			var remaining []model.Player
			for i := range ps {
				if !used[ps[i].ID] {
					remaining = append(remaining, ps[i])
				}
			}
			if bye := chooseBye(remaining); bye != nil {
				// Assign BYE to the selected 'bye' player
				used[bye.ID] = true
				matches = append(matches, model.Match{
					RoundNumber: roundNumber,
					TableNumber: table,
					PlayerA_ID:  bye.ID,
					PlayerB_ID:  ByePlayerID,
					WhiteID:     bye.ID,
					BlackID:     "",
					Result:      "",
				})
				table++
				byeAssigned = true

				if backtrack() {
					return true
				}

				// Undo
				byeAssigned = false
				table--
				matches = matches[:len(matches)-1]
				used[bye.ID] = false
			}
		}

		return false
	}

	if !backtrack() {
		return nil, fmt.Errorf("unable to generate pairings: no rematches and max score difference 1.0 constraints cannot be satisfied")
	}

	return matches, nil
}

const ByePlayerID = "BYE"

// InitializeTournament sets minimal fields and attaches players.
// Title is required; players will be serialized into PlayersData.
// PairingSystem defaults to "SWISS"; ByeScore defaults to 1.0 if unset.
func InitializeTournament(t *model.Tournament, title string, description string, players []model.Player) error {
	// Validate required fields
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("field must be filled: Title is required")
	}
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("field must be filled: Description is required")
	}

	t.Title = title
	t.Description = description
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

	// Locate the target match and round
	var match *model.Match
	var targetRound *model.Round
	for r := range rounds {
		if rounds[r].RoundNumber != roundNumber {
			continue
		}
		targetRound = &rounds[r]
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
		return fmt.Errorf("match not found for round %d, table %d", roundNumber, tableNumber)
	}

	// Validate BYE consistency
	if result == "BYE_A" && match.PlayerB_ID != ByePlayerID {
		return fmt.Errorf("invalid result BYE_A for non-bye match at round %d, table %d", roundNumber, tableNumber)
	}

	// Overwrite match result and scores (supports resubmission safely)
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
		return fmt.Errorf("unknown result %q", result)
	}

	// Check if all matches in this round are now complete
	allComplete := true
	for _, m := range targetRound.Matches {
		if m.Result == "" {
			allComplete = false
			break
		}
	}
	targetRound.IsComplete = allComplete

	// Persist updated rounds before recompute
	if err := t.SetRounds(rounds); err != nil {
		return err
	}

	// Recompute all players (Score, ColorHistory, HasBye, OpponentIDs) from all recorded matches
	if err := RecomputePlayersFromRounds(t); err != nil {
		return err
	}

	// Remove previous MATCH_RESULT_RECORDED event for this round/table to avoid double spending
	events, _ := t.GetEvents()
	filtered := make([]model.Event, 0, len(events))
	for _, e := range events {
		if !(e.Type == "MATCH_RESULT_RECORDED" && e.RoundNumber == roundNumber && e.TableNumber == tableNumber) {
			filtered = append(filtered, e)
		}
	}
	events = filtered

	// Append event: MATCH_RESULT_RECORDED with match snapshot
	detail := struct {
		Match model.Match `json:"match"`
	}{
		Match: *match,
	}
	detailJSON, _ := json.Marshal(detail)
	events = append(events, model.Event{
		EventID:     uuid.New(),
		Type:        "MATCH_RESULT_RECORDED",
		Timestamp:   time.Now(),
		RoundNumber: roundNumber,
		TableNumber: tableNumber,
		Details:     detailJSON,
	})
	if err := t.SetEvents(events); err != nil {
		return err
	}

	// Recompute standings (including Buchholz)
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

// GetStandings returns the players sorted by Score desc, Buchholz desc, Name asc.
// It recomputes Buchholz before sorting to ensure tie-breaks are up-to-date.
func GetStandings(t *model.Tournament) ([]model.Player, error) {
	if err := UpdateStandings(t); err != nil {
		return nil, err
	}
	players, err := t.GetPlayers()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(players, func(i, j int) bool {
		if players[i].Score != players[j].Score {
			return players[i].Score > players[j].Score
		}
		if players[i].Buchholz != players[j].Buchholz {
			return players[i].Buchholz > players[j].Buchholz
		}
		return players[i].Name < players[j].Name
	})
	return players, nil
}

// AdvanceToNextRound runs the pairing engine for the next round and persists the round.
// It updates CurrentRound and TotalPlayers on the tournament.
func AdvanceToNextRound(t *model.Tournament, engine PairingEngine) error {
	players, err := t.GetPlayers()
	if err != nil {
		return err
	}

	// Prevent advancing if the current round exists and is not complete
	if t.CurrentRound > 0 {
		rounds, err2 := t.GetRounds()
		if err2 != nil {
			return err2
		}

		for _, r := range rounds {
			if r.RoundNumber == t.CurrentRound {
				if !r.IsComplete {
					// Collect detailed information about incomplete matches
					var incompleteMatches []string
					var totalMatches int
					var completedMatches int

					for _, m := range r.Matches {
						totalMatches++
						if m.Result == "" {
							// Format player names for better readability
							playerAName := getPlayerName(players, m.PlayerA_ID)
							playerBName := getPlayerName(players, m.PlayerB_ID)

							if m.PlayerB_ID == ByePlayerID {
								incompleteMatches = append(incompleteMatches,
									fmt.Sprintf("Table %d: %s (BYE)", m.TableNumber, playerAName))
							} else {
								incompleteMatches = append(incompleteMatches,
									fmt.Sprintf("Table %d: %s vs %s", m.TableNumber, playerAName, playerBName))
							}
						} else {
							completedMatches++
						}
					}

					// Build detailed error message
					errorMsg := fmt.Sprintf("Cannot advance: Round %d is not complete (%d/%d matches finished).\n",
						t.CurrentRound, completedMatches, totalMatches)

					if len(incompleteMatches) > 0 {
						errorMsg += "Incomplete matches:\n"
						for _, match := range incompleteMatches {
							errorMsg += "• " + match + "\n"
						}
						// Remove trailing newline
						errorMsg = strings.TrimSuffix(errorMsg, "\n")
					}

					return fmt.Errorf("%s", errorMsg)
				}
				break
			}
		}
	}

	nextRoundNumber := t.CurrentRound + 1

	// Pass the tournament to the pairing engine for context
	matches, err := engine.GeneratePairings(t, players, nextRoundNumber)
	if err != nil {
		return err
	}

	// Reorder matches so the previous table-1 winner stays on table 1,
	// BYE (if any) moves to last, and remaining matches follow standings.
	// This prioritizes keeping table over keeping color.
	// Helper: find previous round table-1 winner
	prevTable1Winner := ""
	if t.CurrentRound > 0 {
		if rounds, rErr := t.GetRounds(); rErr == nil {
			for _, r := range rounds {
				if r.RoundNumber == t.CurrentRound {
					for _, m := range r.Matches {
						if m.TableNumber != 1 {
							continue
						}
						switch m.Result {
						case "A_WIN", "BYE_A":
							prevTable1Winner = m.PlayerA_ID
						case "B_WIN":
							prevTable1Winner = m.PlayerB_ID
						default:
							prevTable1Winner = "" // DRAW or empty result: no anchor
						}
						break
					}
					break
				}
			}
		}
	}
	// Build standings rank map for fallback ordering
	rank := map[string]int{}
	if standings, sErr := GetStandings(t); sErr == nil {
		for i := range standings {
			rank[standings[i].ID] = i // smaller index => higher rank
		}
	}
	// Helper: best rank involved in a match (BYE considered worst so it goes last)
	bestRank := func(m model.Match) int {
		if m.PlayerA_ID == ByePlayerID || m.PlayerB_ID == ByePlayerID {
			return len(players) + 1
		}
		ra := rank[m.PlayerA_ID]
		rb := rank[m.PlayerB_ID]
		if ra < rb {
			return ra
		}
		return rb
	}
	hasBye := func(m model.Match) bool {
		return m.PlayerA_ID == ByePlayerID || m.PlayerB_ID == ByePlayerID
	}
	contains := func(m model.Match, id string) bool {
		return id != "" && (m.PlayerA_ID == id || m.PlayerB_ID == id)
	}

	// Sort with priority:
	// 1) match containing previous table-1 winner comes first
	// 2) BYE matches go last
	// 3) remaining matches ordered by standings (bestRank)
	sort.SliceStable(matches, func(i, j int) bool {
		iHasAnchor := contains(matches[i], prevTable1Winner)
		jHasAnchor := contains(matches[j], prevTable1Winner)
		if iHasAnchor != jHasAnchor {
			return iHasAnchor
		}
		iBye := hasBye(matches[i])
		jBye := hasBye(matches[j])
		if iBye != jBye {
			return !iBye
		}
		return bestRank(matches[i]) < bestRank(matches[j])
	})
	// Reassign table numbers after sorting
	for i := range matches {
		matches[i].TableNumber = i + 1
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

// AddPlayer adds a new player to the tournament with an auto-generated UUID.
// Returns the generated player ID and an error if the tournament has already started.
func AddPlayer(t *model.Tournament, name string, club string) (string, error) {
	// Validate required fields
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("player name is required")
	}

	// Prevent adding players after tournament has started
	if t.CurrentRound > 0 {
		return "", fmt.Errorf("cannot add players after tournament has started (current round: %d)", t.CurrentRound)
	}

	// Get current players
	players, err := t.GetPlayers()
	if err != nil {
		return "", err
	}

	// Generate new UUID for the player
	playerID := uuid.NewString()

	// Create new player with initialized fields
	newPlayer := model.Player{
		ID:           playerID,
		Name:         name,
		Score:        0.0,
		OpponentIDs:  []string{},
		Buchholz:     0.0,
		ColorHistory: "",
		HasBye:       false,
		Club:         club,
	}

	// Add the new player
	players = append(players, newPlayer)

	// Update tournament
	if err := t.SetPlayers(players); err != nil {
		return "", err
	}

	// Update total players count
	t.TotalPlayers = len(players)

	return playerID, nil
}

// RecomputePlayersFromRounds rebuilds all player aggregates from the source of truth (rounds).
// This prevents double-counting when results are modified or resubmitted.
func RecomputePlayersFromRounds(t *model.Tournament) error {
	players, err := t.GetPlayers()
	if err != nil {
		return err
	}
	rounds, err := t.GetRounds()
	if err != nil {
		return err
	}

	// Index players by ID for fast updates
	index := make(map[string]*model.Player, len(players))
	for i := range players {
		p := &players[i]
		// Reset aggregate fields
		p.Score = 0
		p.ColorHistory = ""
		p.HasBye = false
		p.OpponentIDs = []string{}
		index[p.ID] = p
	}

	// Apply contributions from all matches that have a recorded result
	// BUT ONLY from rounds <= current round
	for _, r := range rounds {
		// Skip rounds after current round
		if r.RoundNumber > t.CurrentRound {
			continue
		}
		
		for _, m := range r.Matches {
			if m.Result == "" {
				continue
			}

			// Score updates
			if a, ok := index[m.PlayerA_ID]; ok {
				a.Score += m.ScoreA
			}
			if m.PlayerB_ID != ByePlayerID {
				if b, ok := index[m.PlayerB_ID]; ok {
					b.Score += m.ScoreB
				}
			}

			// Opponents and color history
			if m.PlayerB_ID != ByePlayerID {
				// A opponent list + color
				if a, ok := index[m.PlayerA_ID]; ok {
					ensureOpponent(a, m.PlayerB_ID)
					if m.WhiteID == a.ID {
						a.ColorHistory += "W"
					} else if m.BlackID == a.ID {
						a.ColorHistory += "B"
					}
				}
				// B opponent list + color
				if b, ok := index[m.PlayerB_ID]; ok {
					ensureOpponent(b, m.PlayerA_ID)
					if m.WhiteID == b.ID {
						b.ColorHistory += "W"
					} else if m.BlackID == b.ID {
						b.ColorHistory += "B"
					}
				}
			} else {
				// BYE: mark HasBye
				if a, ok := index[m.PlayerA_ID]; ok {
					a.HasBye = true
				}
			}
		}
	}

	// Persist rebuilt players
	return t.SetPlayers(players)
}

// getPlayerName returns the player name for a given ID, or the ID if not found
func getPlayerName(players []model.Player, playerID string) string {
	if playerID == ByePlayerID {
		return "BYE"
	}

	for _, p := range players {
		if p.ID == playerID {
			return p.Name
		}
	}

	// Fallback to ID if name not found
	return playerID
}

// ClearMatchResult clears the result of a specific match in a round
func ClearMatchResult(t *model.Tournament, roundNumber int, tableNumber int) error {
	rounds, err := t.GetRounds()
	if err != nil {
		return err
	}

	// Find the target match and round
	var match *model.Match
	var targetRound *model.Round
	for r := range rounds {
		if rounds[r].RoundNumber != roundNumber {
			continue
		}
		targetRound = &rounds[r]
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
		return fmt.Errorf("match not found for round %d, table %d", roundNumber, tableNumber)
	}

	// Clear the match result
	match.Result = ""
	match.ScoreA = 0.0
	match.ScoreB = 0.0

	// Check if all matches in this round are now incomplete
	allComplete := true
	for _, m := range targetRound.Matches {
		if m.Result == "" {
			allComplete = false
			break
		}
	}
	targetRound.IsComplete = allComplete

	// Persist updated rounds
	if err := t.SetRounds(rounds); err != nil {
		return err
	}

	// Recompute all players from remaining results
	if err := RecomputePlayersFromRounds(t); err != nil {
		return err
	}

	// Recompute standings
	UpdateStandings(t)

	return nil
}

// ClearAllResultsInRound clears all results in a specific round
func ClearAllResultsInRound(t *model.Tournament, roundNumber int) error {
	rounds, err := t.GetRounds()
	if err != nil {
		return err
	}

	// Find the target round
	var targetRound *model.Round
	for r := range rounds {
		if rounds[r].RoundNumber == roundNumber {
			targetRound = &rounds[r]
			break
		}
	}
	if targetRound == nil {
		return fmt.Errorf("round %d not found", roundNumber)
	}

	// Clear all match results in this round
	for m := range targetRound.Matches {
		targetRound.Matches[m].Result = ""
		targetRound.Matches[m].ScoreA = 0.0
		targetRound.Matches[m].ScoreB = 0.0
	}
	targetRound.IsComplete = false

	// Persist updated rounds
	if err := t.SetRounds(rounds); err != nil {
		return err
	}

	// Recompute all players from remaining results
	if err := RecomputePlayersFromRounds(t); err != nil {
		return err
	}

	// Recompute standings
	UpdateStandings(t)

	return nil
}

// GoBackToPreviousRound allows going back to previous round while keeping all results
func GoBackToPreviousRound(t *model.Tournament) error {
	fmt.Printf("DEBUG: GoBackToPreviousRound called - Current round: %d\n", t.CurrentRound)
	
	if t.CurrentRound <= 1 {
		fmt.Printf("DEBUG: Cannot go back - already at round 1 or no rounds exist\n")
		return fmt.Errorf("cannot go back: already at round 1 or no rounds exist (current round: %d)", t.CurrentRound)
	}

	rounds, err := t.GetRounds()
	if err != nil {
		fmt.Printf("DEBUG: Error getting rounds: %v\n", err)
		return err
	}
	
	fmt.Printf("DEBUG: Found %d rounds\n", len(rounds))

	// Check if previous round exists
	previousRoundExists := false
	for _, r := range rounds {
		fmt.Printf("DEBUG: Checking round %d\n", r.RoundNumber)
		if r.RoundNumber == t.CurrentRound-1 {
			previousRoundExists = true
			break
		}
	}

	if !previousRoundExists {
		fmt.Printf("DEBUG: Previous round %d not found\n", t.CurrentRound-1)
		return fmt.Errorf("previous round %d not found", t.CurrentRound-1)
	}

	fmt.Printf("DEBUG: Going back from round %d to round %d\n", t.CurrentRound, t.CurrentRound-1)
	
	// Simply decrement current round - keep all rounds data intact
	t.CurrentRound--

	// Recompute all players from remaining results to ensure consistency
	fmt.Printf("DEBUG: Recomputing players from rounds\n")
	if err := RecomputePlayersFromRounds(t); err != nil {
		fmt.Printf("DEBUG: Error recomputing players: %v\n", err)
		return err
	}

	// Recompute standings
	fmt.Printf("DEBUG: Updating standings\n")
	UpdateStandings(t)

	// Add event log
	events, _ := t.GetEvents()
	detail := struct {
		PreviousRound int    `json:"previous_round"`
		NewRound      int    `json:"new_round"`
		Reason        string `json:"reason"`
	}{
		PreviousRound: t.CurrentRound + 1,
		NewRound:      t.CurrentRound,
		Reason:        "Went back to previous round",
	}
	detailJSON, _ := json.Marshal(detail)
	events = append(events, model.Event{
		EventID:     uuid.New(),
		Type:        "ROUND_REVERTED",
		Timestamp:   time.Now(),
		RoundNumber: t.CurrentRound,
		TableNumber: 0,
		Details:     detailJSON,
	})
	if err := t.SetEvents(events); err != nil {
		fmt.Printf("DEBUG: Error setting events: %v\n", err)
		return err
	}

	fmt.Printf("DEBUG: GoBackToPreviousRound completed successfully - New current round: %d\n", t.CurrentRound)
	return nil
}

// CancelCurrentRound reverts the tournament to the previous round state.
// This removes the current round's pairings and decrements CurrentRound.
// Can only be used if the current round has no recorded results.
func CancelCurrentRound(t *model.Tournament) error {
	if t.CurrentRound <= 0 {
		return fmt.Errorf("cannot cancel: no rounds to cancel (current round: %d)", t.CurrentRound)
	}

	rounds, err := t.GetRounds()
	if err != nil {
		return err
	}

	// Find the current round
	var currentRoundIndex = -1
	for i, r := range rounds {
		if r.RoundNumber == t.CurrentRound {
			currentRoundIndex = i
			break
		}
	}

	if currentRoundIndex == -1 {
		return fmt.Errorf("current round %d not found in rounds data", t.CurrentRound)
	}

	currentRound := rounds[currentRoundIndex]

	// Check if current round has any recorded results
	for _, m := range currentRound.Matches {
		if m.Result != "" {
			return fmt.Errorf("cannot cancel round %d: matches have recorded results. Please clear all results first", t.CurrentRound)
		}
	}

	// Remove the current round from rounds slice
	rounds = append(rounds[:currentRoundIndex], rounds[currentRoundIndex+1:]...)

	// Persist updated rounds
	if err := t.SetRounds(rounds); err != nil {
		return err
	}

	// Decrement current round
	t.CurrentRound--

	// Add event log for cancellation
	events, _ := t.GetEvents()
	detail := struct {
		CancelledRound int    `json:"cancelled_round"`
		Reason         string `json:"reason"`
	}{
		CancelledRound: currentRound.RoundNumber,
		Reason:         "Round cancelled and reverted",
	}
	detailJSON, _ := json.Marshal(detail)
	events = append(events, model.Event{
		EventID:     uuid.New(),
		Type:        "ROUND_CANCELLED",
		Timestamp:   time.Now(),
		RoundNumber: currentRound.RoundNumber,
		TableNumber: 0, // Not applicable for round-level events
		Details:     detailJSON,
	})
	if err := t.SetEvents(events); err != nil {
		return err
	}

	return nil
}

// ExportRoundPairingsToPDF generates a PDF file with tournament round pairings
// Returns the PDF bytes and any error encountered
func ExportRoundPairingsToPDF(t *model.Tournament, roundNumber int) ([]byte, error) {
	// Get tournament data
	players, err := t.GetPlayers()
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	rounds, err := t.GetRounds()
	if err != nil {
		return nil, fmt.Errorf("failed to get rounds: %w", err)
	}

	// Find the specified round
	var targetRound *model.Round
	for _, r := range rounds {
		if r.RoundNumber == roundNumber {
			targetRound = &r
			break
		}
	}

	if targetRound == nil {
		return nil, fmt.Errorf("round %d not found", roundNumber)
	}

	// Create player lookup map for scores
	playerMap := make(map[string]model.Player)
	for _, p := range players {
		playerMap[p.ID] = p
	}

	// Create PDF configuration
	cfg := config.NewBuilder().
		WithPageNumber().
		Build()

	// Create maroto instance
	m := maroto.New(cfg)

	// Add title
	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Tournament: %s", t.Title), props.Text{
					Top:   3,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  16,
				}),
			),
		),
	)

	// Add tournament ID
	m.AddRows(
		row.New(10).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Tournament ID: %s", t.ID.String()), props.Text{
					Top:   2,
					Align: align.Center,
					Size:  10,
				}),
			),
		),
	)

	// Add round number
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Round %d", roundNumber), props.Text{
					Top:   3,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  14,
				}),
			),
		),
	)

	// Add table headers
	m.AddRows(
		row.New(12).Add(
			col.New(2).Add(
				text.New("Table", props.Text{
					Top:   2,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  10,
				}),
			),
			col.New(3).Add(
				text.New("White Player", props.Text{
					Top:   2,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  10,
				}),
			),
			col.New(2).Add(
				text.New("White Points", props.Text{
					Top:   2,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  10,
				}),
			),
			col.New(3).Add(
				text.New("Black Player", props.Text{
					Top:   2,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  10,
				}),
			),
			col.New(2).Add(
				text.New("Black Points", props.Text{
					Top:   2,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  10,
				}),
			),
		),
	)

	// Sort matches by table number
	matches := make([]model.Match, len(targetRound.Matches))
	copy(matches, targetRound.Matches)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].TableNumber < matches[j].TableNumber
	})

	// Add match data rows
	for _, match := range matches {
		whitePlayer := getPlayerName(players, match.WhiteID)
		blackPlayer := getPlayerName(players, match.BlackID)

		// Handle BYE matches
		if match.PlayerB_ID == ByePlayerID {
			blackPlayer = "BYE"
		}

		// Get current points for players
		whitePoints := "0.0"
		blackPoints := "0.0"

		if p, exists := playerMap[match.WhiteID]; exists {
			whitePoints = fmt.Sprintf("%.1f", p.Score)
		}

		if match.BlackID != "" && match.PlayerB_ID != ByePlayerID {
			if p, exists := playerMap[match.BlackID]; exists {
				blackPoints = fmt.Sprintf("%.1f", p.Score)
			}
		} else if match.PlayerB_ID == ByePlayerID {
			blackPoints = "-"
		}

		m.AddRows(
			row.New(8).Add(
				col.New(2).Add(
					text.New(fmt.Sprintf("%d", match.TableNumber), props.Text{
						Top:   1,
						Align: align.Center,
						Size:  9,
					}),
				),
				col.New(3).Add(
					text.New(whitePlayer, props.Text{
						Top:   1,
						Align: align.Left,
						Size:  9,
					}),
				),
				col.New(2).Add(
					text.New(whitePoints, props.Text{
						Top:   1,
						Align: align.Center,
						Size:  9,
					}),
				),
				col.New(3).Add(
					text.New(blackPlayer, props.Text{
						Top:   1,
						Align: align.Left,
						Size:  9,
					}),
				),
				col.New(2).Add(
					text.New(blackPoints, props.Text{
						Top:   1,
						Align: align.Center,
						Size:  9,
					}),
				),
			),
		)
	}

	// Add footer with generation timestamp
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")), props.Text{
					Top:   5,
					Align: align.Center,
					Size:  8,
				}),
			),
		),
	)

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return document.GetBytes(), nil
}

// ExportAllRoundsPairingsToPDF generates a PDF file with all tournament rounds pairings
func ExportAllRoundsPairingsToPDF(t *model.Tournament) ([]byte, error) {
	rounds, err := t.GetRounds()
	if err != nil {
		return nil, fmt.Errorf("failed to get rounds: %w", err)
	}

	if len(rounds) == 0 {
		return nil, fmt.Errorf("no rounds found in tournament")
	}

	players, err := t.GetPlayers()
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	// Create player lookup map
	playerMap := make(map[string]model.Player)
	for _, p := range players {
		playerMap[p.ID] = p
	}

	// Create PDF configuration
	cfg := config.NewBuilder().
		WithPageNumber().
		Build()

	m := maroto.New(cfg)

	// Add main title
	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Tournament: %s - All Rounds", t.Title), props.Text{
					Top:   3,
					Style: fontstyle.Bold,
					Align: align.Center,
					Size:  16,
				}),
			),
		),
	)

	// Add tournament ID
	m.AddRows(
		row.New(10).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Tournament ID: %s", t.ID.String()), props.Text{
					Top:   2,
					Align: align.Center,
					Size:  10,
				}),
			),
		),
	)

	// Process each round
	for i, round := range rounds {
		if i > 0 {
			// Add page break between rounds (except for first round)
			m.AddRows(row.New(10))
		}

		// Add round title
		m.AddRows(
			row.New(15).Add(
				col.New(12).Add(
					text.New(fmt.Sprintf("Round %d", round.RoundNumber), props.Text{
						Top:   3,
						Style: fontstyle.Bold,
						Align: align.Center,
						Size:  14,
					}),
				),
			),
		)

		// Add table headers
		m.AddRows(
			row.New(12).Add(
				col.New(2).Add(
					text.New("Table", props.Text{
						Top:   2,
						Style: fontstyle.Bold,
						Align: align.Center,
						Size:  10,
					}),
				),
				col.New(3).Add(
					text.New("White Player", props.Text{
						Top:   2,
						Style: fontstyle.Bold,
						Align: align.Center,
						Size:  10,
					}),
				),
				col.New(2).Add(
					text.New("White Points", props.Text{
						Top:   2,
						Style: fontstyle.Bold,
						Align: align.Center,
						Size:  10,
					}),
				),
				col.New(3).Add(
					text.New("Black Player", props.Text{
						Top:   2,
						Style: fontstyle.Bold,
						Align: align.Center,
						Size:  10,
					}),
				),
				col.New(2).Add(
					text.New("Black Points", props.Text{
						Top:   2,
						Style: fontstyle.Bold,
						Align: align.Center,
						Size:  10,
					}),
				),
			),
		)

		// Sort matches by table number
		matches := make([]model.Match, len(round.Matches))
		copy(matches, round.Matches)
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].TableNumber < matches[j].TableNumber
		})

		// Add match data rows
		for _, match := range matches {
			whitePlayer := getPlayerName(players, match.WhiteID)
			blackPlayer := getPlayerName(players, match.BlackID)

			if match.PlayerB_ID == ByePlayerID {
				blackPlayer = "BYE"
			}

			whitePoints := "0.0"
			blackPoints := "0.0"

			if p, exists := playerMap[match.WhiteID]; exists {
				whitePoints = fmt.Sprintf("%.1f", p.Score)
			}

			if match.BlackID != "" && match.PlayerB_ID != ByePlayerID {
				if p, exists := playerMap[match.BlackID]; exists {
					blackPoints = fmt.Sprintf("%.1f", p.Score)
				}
			} else if match.PlayerB_ID == ByePlayerID {
				blackPoints = "-"
			}

			m.AddRows(
				row.New(8).Add(
					col.New(2).Add(
						text.New(fmt.Sprintf("%d", match.TableNumber), props.Text{
							Top:   1,
							Align: align.Center,
							Size:  9,
						}),
					),
					col.New(3).Add(
						text.New(whitePlayer, props.Text{
							Top:   1,
							Align: align.Left,
							Size:  9,
						}),
					),
					col.New(2).Add(
						text.New(whitePoints, props.Text{
							Top:   1,
							Align: align.Center,
							Size:  9,
						}),
					),
					col.New(3).Add(
						text.New(blackPlayer, props.Text{
							Top:   1,
							Align: align.Left,
							Size:  9,
						}),
					),
					col.New(2).Add(
						text.New(blackPoints, props.Text{
							Top:   1,
							Align: align.Center,
							Size:  9,
						}),
					),
				),
			)
		}

		// Add spacing between rounds
		if i < len(rounds)-1 {
			m.AddRows(row.New(10))
		}
	}

	// Add footer
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")), props.Text{
					Top:   5,
					Align: align.Center,
					Size:  8,
				}),
			),
		),
	)

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return document.GetBytes(), nil
}
