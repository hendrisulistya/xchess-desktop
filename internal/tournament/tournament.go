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
		if ps[i].Rating != ps[j].Rating {
			return ps[i].Rating > ps[j].Rating
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
	roundIndex := -1
	for r := range rounds {
		if rounds[r].RoundNumber != roundNumber {
			continue
		}
		for m := range rounds[r].Matches {
			if rounds[r].Matches[m].TableNumber == tableNumber {
				match = &rounds[r].Matches[m]
				roundIndex = r
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

	// Prevent double-counting: if this match already has a result, do not apply again.
	if match.Result != "" {
		return fmt.Errorf("match result already recorded for round %d, table %d", roundNumber, tableNumber)
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
			if match.WhiteID == a.ID {
				a.ColorHistory += "W"
			} else if match.BlackID == a.ID {
				a.ColorHistory += "B"
			}
		} else {
			a.HasBye = true
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

	// Mark round complete if all matches have recorded results
	if roundIndex >= 0 {
		complete := true
		for _, m := range rounds[roundIndex].Matches {
			if m.Result == "" {
				complete = false
				break
			}
		}
		rounds[roundIndex].IsComplete = complete
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

// GetStandings returns the players sorted by Score desc, Buchholz desc, Rating desc, Name asc.
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
		if players[i].Rating != players[j].Rating {
			return players[i].Rating > players[j].Rating
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
					return fmt.Errorf("cannot advance: current round %d is not complete", t.CurrentRound)
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
