package ges

import (
	"time"
)

// Snapshot represents the current persisted state of an aggregate
// at a specific version, optionally loaded from storage.
type Snapshot struct {
	State   any       // The deserialized state
	Version int64     // Aggregate version at which the snapshot was taken
	Found   bool      // Whether a snapshot exists
	At      time.Time // Timestamp of when it was taken
}
