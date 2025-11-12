package pgx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mickamy/go-event-sourcing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventStore is a concrete EventStore backed by PostgreSQL (pgx).
// It supports optimistic concurrency, JSON-encoded payloads, and optional
// context-derived Metadata injection via a user-supplied MetadataExtractor.
type EventStore struct {
	pool         *pgxpool.Pool
	typeRegistry map[string]ges.EventCodec
	extractor    ges.MetadataExtractor
}

// Option configures EventStore.
type Option func(*EventStore)

// WithTypeRegistry sets the registry that maps event type names to codecs.
func WithTypeRegistry(reg map[string]ges.EventCodec) Option {
	return func(s *EventStore) { s.typeRegistry = reg }
}

// WithMetadataExtractor sets a function that builds Metadata from context.
// When provided, Append() will merge extracted metadata with the explicit md;
// explicit keys take precedence over extracted ones.
func WithMetadataExtractor(ex ges.MetadataExtractor) Option {
	return func(s *EventStore) { s.extractor = ex }
}

// NewEventStore creates a Postgres-backed EventStore.
func NewEventStore(pool *pgxpool.Pool, opts ...Option) *EventStore {
	s := &EventStore{
		pool:         pool,
		typeRegistry: map[string]ges.EventCodec{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Append persists a batch of events using optimistic concurrency control.
func (s *EventStore) Append(
	ctx context.Context,
	streamID string,
	expectedVersion int64,
	events []ges.Event,
	md ges.Metadata,
) (int64, error) {
	// Merge context-derived metadata (if configured) with explicit md.
	// Later maps take precedence → explicit md overrides extracted.
	if s.extractor != nil {
		extracted := s.extractor(ctx)
		md = extracted.Merge(md)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("ges-pgx: could not begin transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	// Read current stream version.
	var currentVersion int64
	if err := tx.QueryRow(
		ctx,
		`SELECT COALESCE(MAX(version), 0) FROM events WHERE stream_id = $1`,
		streamID,
	).Scan(&currentVersion); err != nil {
		return 0, fmt.Errorf("ges-pgx: could not get current version: %w", err)
	}
	if currentVersion != expectedVersion {
		return 0, &ges.VersionConflictError{
			StreamID:        streamID,
			ExpectedVersion: expectedVersion,
			ActualVersion:   currentVersion,
		}
	}

	if len(events) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return 0, fmt.Errorf("ges-pgx: could not commit transaction: %w", err)
		}
		return expectedVersion, nil
	}

	// Insert each event with the next version.
	for _, e := range events {
		eventType := ges.EventType(e)
		codec := s.typeRegistry[eventType]
		if codec == nil {
			return 0, fmt.Errorf("ges-pgx: no codec registered for event type %q", eventType)
		}

		payload, err := codec.Encode(e)
		if err != nil {
			return 0, fmt.Errorf("ges-pgx: could not encode event: %w", err)
		}

		meta, err := json.Marshal(md)
		if err != nil {
			return 0, fmt.Errorf("ges-pgx: could not encode metadata: %w", err)
		}

		currentVersion++

		if _, err := tx.Exec(
			ctx,
			`
			INSERT INTO events (stream_id, version, event_type, payload, metadata)
			VALUES ($1, $2, $3, $4, $5)
			`,
			streamID,
			currentVersion,
			eventType,
			payload,
			meta,
		); err != nil {
			if isUniqueViolation(err) {
				return 0, &ges.VersionConflictError{
					StreamID:        streamID,
					ExpectedVersion: expectedVersion,
					ActualVersion:   currentVersion,
				}
			}
			return 0, fmt.Errorf("ges-pgx: could not insert event: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("ges-pgx: could not commit transaction: %w", err)
	}
	return currentVersion, nil
}

// Load returns all events for a given stream strictly after fromVersion,
// ordered by version ascending. The second return value is the last version read.
func (s *EventStore) Load(
	ctx context.Context,
	streamID string,
	fromVersion int64,
) ([]ges.Event, int64, error) {
	rows, err := s.pool.Query(
		ctx,
		`
		SELECT version, event_type, payload
		FROM events
		WHERE stream_id = $1 AND version > $2
		ORDER BY version ASC
		`,
		streamID,
		fromVersion,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("ges-pgx: could not query events: %w", err)
	}
	defer rows.Close()

	var out []ges.Event
	var last int64

	for rows.Next() {
		var version int64
		var eventType string
		var payload []byte

		if err := rows.Scan(&version, &eventType, &payload); err != nil {
			return nil, 0, fmt.Errorf("ges-pgx: could not scan event: %w", err)
		}

		codec := s.typeRegistry[eventType]
		if codec == nil {
			return nil, 0, fmt.Errorf("unknown event type: %s", eventType)
		}

		ev, err := codec.Decode(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("ges-pgx: could not decode event: %w", err)
		}

		out = append(out, ev)
		last = version
	}
	return out, last, nil
}

// SaveSnapshot upserts the snapshot state for a stream at a given version.
// Snapshots are an optimization for fast rehydration and are safe to treat
// as a cache—failure to save should not compromise domain consistency.
func (s *EventStore) SaveSnapshot(
	ctx context.Context,
	streamID string,
	version int64,
	state any,
) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(
		ctx,
		`
		INSERT INTO snapshots (stream_id, version, state)
		VALUES ($1, $2, $3)
		ON CONFLICT (stream_id) DO UPDATE
		SET version = EXCLUDED.version,
		    state   = EXCLUDED.state
		`,
		streamID,
		version,
		data,
	)
	return err
}

// LoadSnapshot retrieves the latest snapshot for a stream. If not found, Found=false.
// The State is returned as a generic structure (typically map[string]any) since the
// library does not enforce a concrete aggregate type; applications can re-decode it.
func (s *EventStore) LoadSnapshot(
	ctx context.Context,
	streamID string,
) (ges.Snapshot, error) {
	row := s.pool.QueryRow(
		ctx,
		`SELECT version, state, at FROM snapshots WHERE stream_id = $1`,
		streamID,
	)

	var version int64
	var raw []byte
	var at time.Time

	if err := row.Scan(&version, &raw, &at); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ges.Snapshot{Found: false}, nil
		}
		return ges.Snapshot{}, fmt.Errorf("ges-pgx: could not scan snapshot: %w", err)
	}

	// Decode into a generic map by default; callers may re-decode to a concrete type.
	var state map[string]any
	if err := json.Unmarshal(raw, &state); err != nil {
		return ges.Snapshot{}, fmt.Errorf("ges-pgx: could not unmarshal snapshot: %w", err)
	}

	return ges.Snapshot{
		State:   state,
		Version: version,
		Found:   true,
		At:      at,
	}, nil
}

var _ ges.EventStore = (*EventStore)(nil)
