package ges

import (
	"fmt"
)

var (
	// ErrVersionConflict indicates that the expectedVersion did not match
	// the current version in the store, typically due to concurrent writes.
	ErrVersionConflict = fmt.Errorf("eventstore: version conflict")
)

// VersionConflictError provides structured information about version mismatch.
type VersionConflictError struct {
	StreamID        string
	ExpectedVersion int64
	ActualVersion   int64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("version conflict on stream %s: expected=%d actual=%d", e.StreamID, e.ExpectedVersion, e.ActualVersion)
}

// Is allows errors.Is(err, ErrVersionConflict) to match this type.
func (e *VersionConflictError) Is(target error) bool {
	return target == ErrVersionConflict
}
