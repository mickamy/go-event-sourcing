package ges

// Base is an embeddable helper to implement Aggregate boilerplate.
// Semantics:
//   - Apply(e): mutate state via applier and bump version by 1. Does NOT enqueue.
//   - Raise(e): Apply(e) + enqueue to pending (for newly produced events).
//   - Version(): current version INCLUDING pending.
//   - Flush(): returns pending and clears it; also returns
//     expectedVersion = currentVersion - len(pending_before).
type Base struct {
	id      string
	version int64
	pending []Event
	applier func(Event)
}

// Init sets the stream ID and the state mutation function (applier).
func (b *Base) Init(streamID string, applier func(Event)) {
	b.id = streamID
	b.applier = applier
}

// StreamID returns the unique identifier for this aggregate’s event stream.
func (b *Base) StreamID() string { return b.id }

// SetStreamID overrides the stream ID (e.g., when the first event assigns it).
func (b *Base) SetStreamID(streamID string) { b.id = streamID }

// SetApplier replaces the state mutation function.
func (b *Base) SetApplier(applier func(Event)) { b.applier = applier }

// SetVersion forces the current version (used when restoring from a snapshot).
// It sets the internal counter; no pending events are affected.
func (b *Base) SetVersion(v int64) { b.version = v }

// Apply mutates state by a single event and advances the version by 1.
// Typically used for event replay (rehydration) or confirming committed events.
func (b *Base) Apply(e Event) {
	if b.applier != nil {
		b.applier(e)
	}
	b.version++
}

// Raise records a new domain event: Apply(e) and enqueue it into the pending buffer.
// Call Flush to obtain and clear pending events for persistence.
func (b *Base) Raise(e Event) {
	b.Apply(e)
	b.pending = append(b.pending, e)
}

// Flush returns all uncommitted events and clears the pending buffer.
// expectedVersion = currentVersion - len(pendingBeforeFlush)
func (b *Base) Flush() (events []Event, expectedVersion int64) {
	events = b.pending
	expectedVersion = b.version - int64(len(events))
	b.pending = nil
	return
}

// Version returns the current aggregate version INCLUDING pending events.
func (b *Base) Version() int64 { return b.version }
