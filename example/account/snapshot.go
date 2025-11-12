package main

import (
	"encoding/json"

	"github.com/mickamy/go-event-sourcing"
)

// AccountSnapshot is the persisted state shape stored in snapshots.
type AccountSnapshot struct {
	ID      string `json:"id"`
	Owner   string `json:"owner"`
	Balance int64  `json:"balance"`
	Version int64  `json:"version"`
}

// serializeState converts the in-memory aggregate into a persistable snapshot.
func serializeState(a *Account) any {
	return AccountSnapshot{
		ID:      a.id,
		Owner:   a.owner,
		Balance: a.balance,
		Version: a.version,
	}
}

func decodeSnapshot(snap ges.Snapshot) (AccountSnapshot, bool, error) {
	if !snap.Found || snap.State == nil {
		return AccountSnapshot{}, false, nil
	}
	// State(map[string]any) → JSON → AccountSnapshot
	raw, err := json.Marshal(snap.State)
	if err != nil {
		return AccountSnapshot{}, false, err
	}
	var out AccountSnapshot
	if err := json.Unmarshal(raw, &out); err != nil {
		return AccountSnapshot{}, false, err
	}
	return out, true, nil
}
