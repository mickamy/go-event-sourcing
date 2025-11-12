package mem

import (
	"context"
	"sync"
	"time"

	"github.com/mickamy/go-event-sourcing"
)

// Store is an in-memory EventStore implementation.
// It is concurrency-safe and suitable for tests, prototypes, and local runs.
// NOTE: Events and snapshots are kept in-process and will be lost on restart.
type Store struct {
	mu        sync.RWMutex
	streams   map[string][]storedEvent
	snapshots map[string]snapshot
	extractor ges.MetadataExtractor
}

type storedEvent struct {
	version  int64
	payload  ges.Event
	metadata ges.Metadata
	typ      string
	at       time.Time
}

type snapshot struct {
	version int64
	state   any
	at      time.Time
}

// Option configures the in-memory Store.
type Option func(*Store)

// WithMetadataExtractor sets a function that builds Metadata from context.
// When provided, Append will merge extracted metadata with the explicit md;
// explicit keys take precedence over extracted ones.
func WithMetadataExtractor(ex ges.MetadataExtractor) Option {
	return func(s *Store) { s.extractor = ex }
}

// New creates a new in-memory Store.
func New(opts ...Option) *Store {
	st := &Store{
		streams:   make(map[string][]storedEvent),
		snapshots: make(map[string]snapshot),
	}
	for _, opt := range opts {
		opt(st)
	}
	return st
}

// Append persists a batch of events using optimistic concurrency control.
//
// Semantics:
//   - expectedVersion must equal the current persisted version for streamID.
//   - On version mismatch, returns *ges.VersionConflictError (errors.Is with ErrVersionConflict works).
//   - Returns the new current version after successful append.
//   - If events is empty, it acts as a pure version check and returns expectedVersion.
func (s *Store) Append(
	ctx context.Context,
	streamID string,
	expectedVersion int64,
	events []ges.Event,
	md ges.Metadata,
) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Merge context-derived metadata (if configured) with explicit md.
	// Later maps take precedence → explicit md overrides extracted.
	if s.extractor != nil {
		extracted := s.extractor(ctx)
		md = extracted.Merge(md)
	}

	seq := s.streams[streamID]
	currentVersion := int64(len(seq))
	if currentVersion != expectedVersion {
		return 0, &ges.VersionConflictError{
			StreamID:        streamID,
			ExpectedVersion: expectedVersion,
			ActualVersion:   currentVersion,
		}
	}

	if len(events) == 0 {
		// Nothing to append; treat as a successful check.
		return expectedVersion, nil
	}

	now := time.Now()
	// Append each event, assigning the next version number.
	for _, e := range events {
		currentVersion++
		seq = append(seq, storedEvent{
			version:  currentVersion,
			payload:  e,
			metadata: md, // already a new map via Merge; safe to reuse
			typ:      ges.EventType(e),
			at:       now,
		})
	}
	s.streams[streamID] = seq
	return currentVersion, nil
}

// Load returns all events for a given stream strictly after fromVersion,
// ordered by version ascending. The second return value is the last version read.
func (s *Store) Load(
	_ context.Context,
	streamID string,
	fromVersion int64,
) ([]ges.Event, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seq := s.streams[streamID]
	if len(seq) == 0 {
		return nil, 0, nil
	}

	// fromVersion is exclusive; indexes are zero-based (version = index+1)
	start := fromVersion
	if start < 0 {
		start = 0
	}
	if start > int64(len(seq)) {
		start = int64(len(seq))
	}

	var out []ges.Event
	for i := start; i < int64(len(seq)); i++ {
		out = append(out, seq[i].payload)
	}
	last := int64(0)
	if len(seq) > 0 {
		last = seq[len(seq)-1].version
	}
	return out, last, nil
}

// SaveSnapshot upserts the snapshot state for a stream at a given version.
// Snapshots are an optimization for fast rehydration and are safe to treat as cache.
func (s *Store) SaveSnapshot(
	_ context.Context,
	streamID string,
	version int64,
	state any,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots[streamID] = snapshot{
		version: version,
		state:   state,
		at:      time.Now(),
	}
	return nil
}

// LoadSnapshot retrieves the latest snapshot for a stream. If not found, Found=false.
// State is returned “as-is” (typically the same shape you saved, e.g., map[string]any).
func (s *Store) LoadSnapshot(
	_ context.Context,
	streamID string,
) (ges.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap, ok := s.snapshots[streamID]
	if !ok {
		return ges.Snapshot{Found: false}, nil
	}
	return ges.Snapshot{
		State:   snap.state,
		Version: snap.version,
		Found:   true,
		At:      snap.at,
	}, nil
}

var _ ges.EventStore = (*Store)(nil)
