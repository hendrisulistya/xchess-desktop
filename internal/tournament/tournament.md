# XChess Tournament Rules and Gameplay Specification

This document defines the tournament rules and describes how they are implemented in code. Use it as the source of truth to update internal/tournament/tournament.go and related code.

## Overview
- System: Swiss pairing
- Rounds: Multiple; each round has matches between players
- Scoring: Win = 1.0, Draw = 0.5, Loss = 0.0, Bye = configurable (default 1.0)
- Tie-break: Buchholz (sum of opponents’ scores, excluding BYE)
- Color tracking: Minimal balancing based on last color played
- Persistence: Players and rounds stored as JSON fields in a single Tournament record

## Data Model (Go)
- Tournament
  - PlayersData: JSON of []Player
  - RoundsData: JSON of []Round
  - CurrentRound: int
  - TotalPlayers: int
  - ByeScore: float64 (default 1.0)
  - PairingSystem: string (default "SWISS")
- Round
  - RoundNumber: int
  - Matches: []Match
  - IsComplete: bool
- Match
  - RoundNumber: int
  - TableNumber: int
  - PlayerA_ID: string
  - PlayerB_ID: string (set to "BYE" for bye)
  - WhiteID: string
  - BlackID: string
  - Result: string ("A_WIN", "B_WIN", "DRAW", "BYE_A")
  - ScoreA, ScoreB: float64
- Player
  - ID, Name
  - Score
  - OpponentIDs: []string
  - Buchholz
  - ColorHistory: string ("W"/"B" appended per match)
  - HasBye: bool
  - Rating: int (optional)

## Lifecycle

1. Initialize Tournament
   - Action: Set metadata and serialize players/rounds
   - Required: Title and Description must be provided; error if either is empty ("field must be filled")
   - Code: InitializeTournament(t, title, description, players) sets:
     - Status = "ACTIVE"
     - CurrentRound = 0
     - TotalPlayers = len(players)
     - PairingSystem = "SWISS" if empty
     - ByeScore = 1.0 if zero
     - PlayersData and RoundsData initialized

2. Advance To Next Round
   - Action: Increment round and generate pairings
   - Required Rule: You must NOT advance if the current round is unfinished
   - Implementation:
     - New round is appended with IsComplete = false
     - Pairings generated via PairingEngine.GeneratePairings(...)
   - Enforcement (required):
     - Before advancing, check if a round exists with RoundNumber == CurrentRound and ensure IsComplete == true
     - If not complete, return an error (e.g., "cannot advance: current round X is not complete")

3. Record Match Result
   - Action: Update match result and scores
   - Codes:
     - "A_WIN": ScoreA=1.0, ScoreB=0.0
     - "B_WIN": ScoreA=0.0, ScoreB=1.0
     - "DRAW": ScoreA=0.5, ScoreB=0.5
     - "BYE_A": ScoreA=ByeScore (default 1.0), ScoreB=0.0; PlayerB_ID should be "BYE"
   - Player updates:
     - Add opponent IDs (skip BYE for opponent updates)
     - Update ColorHistory ("W" if the player is White, "B" if Black)
     - Set HasBye for bye recipients
   - Round completion:
     - After setting a result, mark the round IsComplete = true only if all matches have non-empty Result

4. Standings & Tie-breaks
   - Buchholz: Sum of opponents’ current scores (excluding BYE)
   - Recompute after every recorded result via UpdateStandings(...)

## Pairing Rules

- Round 1
  - Use external swiss-tool to generate random pairings
  - Map internal player IDs to swiss-tool participants
  - Colors: Player A is assigned White; Player B is Black

- Subsequent Rounds
  - Sort players by:
    - Score desc
    - Buchholz desc
    - Rating desc
    - Name asc
  - Pairing constraints:
    - No rematches allowed
    - Maximum score difference between paired players: 1.0
  - Pairing selection:
    - Prefer same-score opponents (within constraints)
    - Otherwise choose closest-score opponents (still within max difference 1.0)
  - Color assignment:
    - Minimal balancing: If a player’s last ColorHistory == 'W', try to give them Black; otherwise keep default
  - Bye policy:
    - If the number of players is odd, assign exactly one BYE
    - Choose bye among unpaired candidates by lowest score, preferring players without prior bye; ties by lower Buchholz, then Name
    - If constraints cannot be satisfied with an even number of players (no rematches and <= 1.0 score difference), pairing fails with an error

## Constants
- ByePlayerID = "BYE"
- Default ByeScore = 1.0 (when t.ByeScore is unset)

## Data Access & History
- GetRounds(t Tournament) -> []Round: Deserializes rounds for the specific tournament instance
- GetPlayers(t Tournament) -> []Player: Deserializes players
- Current round matches:
  - Use t.CurrentRound with GetRounds() and filter by RoundNumber
  - App-level helper: App.GetCurrentRound()

## Implementation Pointers (Where to change in code)
- Pairing behavior and constraints:
  - internal/tournament/tournament.go, SwissToolAdapter.GeneratePairings(...): enforce no rematches and max score difference 1.0 with backtracking
- Scoring and color tracking for results:
  - internal/tournament/tournament.go, RecordMatchResult(...), ensure ColorHistory and HasBye updates
- Round completion gate (must-have):
  - internal/tournament/tournament.go, AdvanceToNextRound(...): guard against advancing when the current round IsComplete == false
- Tie-breaks:
  - internal/tournament/tournament.go, UpdateStandings(...): recompute Buchholz excluding BYE
- Persistence:
  - internal/tournament/tournament.go or model methods: GetPlayers/SetPlayers, GetRounds/SetRounds

## Notes
- All mutations must be serialized back using SetRounds and SetPlayers after updates.
- Ensure frontend uses App.NextRound only when the current round is complete, or handle backend errors gracefully.
- If pairing fails with an even number of players under constraints, consider relaxing constraints in the spec or adjusting participants; backend will return an error rather than violating rules.
- Future extensions can add helper functions for match history retrieval (e.g., per-round or full history) if needed.