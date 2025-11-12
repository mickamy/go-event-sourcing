package storetest

import (
	"errors"
	"testing"

	ges "github.com/mickamy/go-event-sourcing"
)

type Opened struct{ ID string }

func (Opened) EventType() string { return "Opened" }

type Added struct{ N int }

func (Added) EventType() string { return "Added" }

// Factory creates a new EventStore instance for testing.
// Each test should receive a fresh, isolated instance.
// Use t.Cleanup for teardown logic if necessary.
type Factory func(t *testing.T) ges.EventStore

// Registry provides a minimal codec registry used for tests.
// It avoids dependency on domain-specific event definitions.
func Registry() map[string]ges.EventCodec {
	return map[string]ges.EventCodec{
		"Opened": ges.JSONCodec[Opened](),
		"Added":  ges.JSONCodec[Added](),
	}
}

// Run executes a suite of compliance tests that verify an EventStore
// implementation adheres to the expected semantics.
// Each subtest runs in parallel, so stores must be concurrency-safe.
func Run(t *testing.T, newStore Factory) {
	t.Run("append/load/version", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		s := newStore(t)

		streamID := "Stream:1"

		// Append first event
		v, err := s.Append(ctx, streamID, 0, []ges.Event{
			Opened{ID: "1"},
		}, nil)
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}
		if v != 1 {
			t.Fatalf("expected version 1, got %d", v)
		}

		// Append second event
		v, err = s.Append(ctx, streamID, v, []ges.Event{
			Added{N: 5},
		}, nil)
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}
		if v != 2 {
			t.Fatalf("expected version 2, got %d", v)
		}

		// Load all events
		evs, last, err := s.Load(ctx, streamID, 0)
		if err != nil {
			t.Fatalf("load failed: %v", err)
		}
		if len(evs) != 2 {
			t.Fatalf("expected 2 events, got %d", len(evs))
		}
		if last != 2 {
			t.Fatalf("expected last version 2, got %d", last)
		}
	})

	t.Run("version conflict", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		s := newStore(t)
		streamID := "Stream:2"

		// First append succeeds
		if _, err := s.Append(ctx, streamID, 0, []ges.Event{
			Opened{ID: "2"},
		}, nil); err != nil {
			t.Fatalf("append failed: %v", err)
		}

		// Second append with wrong expected version should fail
		_, err := s.Append(ctx, streamID, 0, []ges.Event{
			Added{N: 1},
		}, nil)

		var vc *ges.VersionConflictError
		if !errors.As(err, &vc) {
			t.Fatalf("expected VersionConflictError, got %v", err)
		}
	})
}
