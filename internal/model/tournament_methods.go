package model

import "encoding/json"

// GetPlayers deserializes the PlayersData field into a slice of Player structs.
func (t Tournament) GetPlayers() ([]Player, error) {
	// ... existing code ...
	var players []Player
	if t.PlayersData == nil {
		return players, nil
	}
	err := json.Unmarshal(t.PlayersData, &players)
	return players, err
	// ... existing code ...
}

// SetPlayers serializes a slice of Player structs into the PlayersData field.
func (t *Tournament) SetPlayers(players []Player) error {
	// ... existing code ...
	data, err := json.Marshal(players)
	if err != nil {
		return err
	}
	t.PlayersData = data
	return nil
	// ... existing code ...
}

// GetRounds deserializes the RoundsData field into a slice of Round structs.
func (t Tournament) GetRounds() ([]Round, error) {
	// ... existing code ...
	var rounds []Round
	if t.RoundsData == nil {
		return rounds, nil
	}
	err := json.Unmarshal(t.RoundsData, &rounds)
	return rounds, err
	// ... existing code ...
}

// SetRounds serializes a slice of Round structs into the RoundsData field.
func (t *Tournament) SetRounds(rounds []Round) error {
	// ... existing code ...
	data, err := json.Marshal(rounds)
	if err != nil {
		return err
	}
	t.RoundsData = data
	return nil
	// ... existing code ...
}