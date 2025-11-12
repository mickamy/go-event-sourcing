package ges

import (
	"fmt"
	"time"
)

// Event is a semantic alias of `any` that represents a domain event payload.
type Event any

// StoredEvent represents an event that has been persisted in the event store.
type StoredEvent struct {
	ID       string
	Type     string
	Payload  Event
	Metadata Metadata
	StreamID string
	Version  int64
	At       time.Time
}

// EventType returns the canonical name for a given event.
// If the event implements `EventType() string`, that value is used.
// Otherwise, it falls back to the Go type name (e.g., "account.AccountOpened").
func EventType(e Event) string {
	if named, ok := e.(interface{ EventType() string }); ok {
		return named.EventType()
	}
	return fmt.Sprintf("%T", e)
}
