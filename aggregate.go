package ges

// Aggregate represents a domain entity that is rebuilt from a stream of events.
type Aggregate interface {
	// StreamID returns the unique identifier for this aggregate’s event stream.
	// e.g., "Account:12345"
	StreamID() string

	// Apply mutates the aggregate’s state by applying a single event.
	// It is typically called during event replay (rehydration) or when recording new events.
	Apply(e Event)

	// Flush returns all uncommitted events and clears the pending buffer.
	// It also returns the expected stream version for optimistic locking.
	//
	// expectedVersion = currentVersion - len(pendingBeforeFlush)
	Flush() (events []Event, expectedVersion int64)

	// Version returns the current aggregate version (for optimistic locking).
	Version() int64
}
