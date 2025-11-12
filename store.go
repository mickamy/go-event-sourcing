package ges

import (
	"context"
)

// EventStore defines the interface for persisting and retrieving events
// in an event-sourced system.
//
// It provides the core abstraction that enables aggregates to record domain
// events (via Append) and later rebuild their state (via Load).
//
// Implementations may persist events to PostgreSQL, DynamoDB, or other durable
// backends. All operations must be safe for concurrent use and respect
// optimistic locking semantics.
type EventStore interface {
	// Load returns all events for the given stream starting from a specific version.
	// The returned slice must be ordered by version ascending.
	Load(ctx context.Context, streamID string, fromVersion int64) ([]Event, int64, error)

	// Append writes a batch of events to the store.
	//
	// expectedVersion must match the current persisted version of the stream.
	// If the versions differ (for example, due to a concurrent writer),
	// the method must return a *VersionConflictError, which can be tested with:
	//
	//   if errors.Is(err, ges.ErrVersionConflict) { ... }
	//
	// Implementations must ensure atomicity — either all events are appended,
	// or none are.
	Append(ctx context.Context, streamID string, expectedVersion int64, events []Event, md Metadata) (int64, error)

	// SaveSnapshot stores a serialized representation of the aggregate’s current state.
	// This is an optional optimization to avoid replaying the entire event history
	// when reloading aggregates. Snapshots are safe to treat as caches — failure
	// to save should not affect event consistency.
	SaveSnapshot(ctx context.Context, streamID string, version int64, state any) error

	// LoadSnapshot retrieves the latest snapshot for the given stream.
	//
	// If no snapshot exists, the returned Snapshot will have Found=false
	// and zero values for State and Version.
	LoadSnapshot(ctx context.Context, streamID string) (Snapshot, error)
}
