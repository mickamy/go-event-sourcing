package main

import (
	"context"
	"fmt"

	"github.com/mickamy/go-event-sourcing"
)

// AccountRepository loads and saves Account aggregates using an EventStore.
type AccountRepository struct {
	store ges.EventStore
}

// NewAccountRepository creates a repository backed by the given store.
func NewAccountRepository(store ges.EventStore) *AccountRepository {
	return &AccountRepository{store: store}
}

// Load fetches and rehydrates an Account by its ID.
// It tries a snapshot first, then loads the delta events.
func (r *AccountRepository) Load(ctx context.Context, id string) (*Account, error) {
	streamID := "Account:" + id

	var a Account

	// 1) Try snapshot
	snap, err := r.store.LoadSnapshot(ctx, streamID)
	if err != nil {
		return nil, err
	}
	if s, ok, err := decodeSnapshot(snap); err != nil {
		return nil, err
	} else if ok {
		a.id = s.ID
		a.owner = s.Owner
		a.balance = s.Balance
		a.version = s.Version
		a.opened = s.ID != ""
	}

	// 2) Apply delta events
	evs, last, err := r.store.Load(ctx, streamID, a.Version())
	if err != nil {
		return nil, err
	}
	a.Restore(evs)
	if last != a.Version() {
		return nil, fmt.Errorf(
			"version mismatch after Restore: aggregate=%d, store=%d",
			a.Version(), last,
		)
	}

	return &a, nil
}

// Save persists the aggregate's pending events with optimistic locking.
// On success, it clears pending events.
func (r *AccountRepository) Save(ctx context.Context, a *Account, md ges.Metadata) error {
	evs, expected := a.Flush()
	if len(evs) == 0 {
		return nil
	}
	_, err := r.store.Append(ctx, a.StreamID(), expected, evs, md)
	return err
}
