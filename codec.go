package ges

import (
	"encoding/json"
	"fmt"
)

// EventCodec defines how events are encoded/decoded for persistence.
// Each event type should register its codec in the EventStore.
type EventCodec interface {
	Encode(v any) ([]byte, error)
	Decode(b []byte) (any, error)
}

// JSONCodec is a generic implementation of EventCodec for JSON-based encoding.
func JSONCodec[T any]() EventCodec {
	return jsonCodec[T]{}
}

type jsonCodec[T any] struct{}

func (jsonCodec[T]) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec[T]) Decode(b []byte) (any, error) {
	var v T
	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil, fmt.Errorf("ges: failed to decode json: %w", err)
	}
	return v, err
}
