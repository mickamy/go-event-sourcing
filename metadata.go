package ges

import (
	"context"
)

// Metadata carries contextual information that accompanies events.
// Typical keys include tenant_id, user_id, correlation_id, and trace_id.
type Metadata map[string]any

// Merge returns a new Metadata that combines the receiver with the given maps.
// It is safe to call on a nil receiver. Later maps take precedence over earlier ones.
// The receiver is not modified.
func (m Metadata) Merge(ms ...Metadata) Metadata {
	out := make(Metadata)

	if m != nil {
		for k, v := range m {
			out[k] = v
		}
	}

	for _, other := range ms {
		for k, v := range other {
			out[k] = v
		}
	}
	return out
}

// MetadataExtractor builds Metadata from a context.
// Applications can supply their own extractor that knows about
// private context keys (tenant_id, user_id, correlation_id, trace_id, etc.).
type MetadataExtractor func(ctx context.Context) Metadata
