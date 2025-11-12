# go-event-sourcing

A lightweight, framework-agnostic Event Sourcing library for Go. It provides a minimal interface for persisting, loading, and snapshotting domain events, with pluggable backends such as PostgreSQL (`pgx`) and MySQL.

## Features

* Clean, minimal core (`EventStore`, `Aggregate`, `Event`, `Metadata`)
* Supports optimistic concurrency and version conflict detection
* Snapshot support for efficient rehydration
* Context-based metadata injection (`tenant_id`, `user_id`, etc.)
* Pluggable backends (`stores/pgx`, `stores/mysql`)
* No external dependencies in the core package

## Installation

```bash
go get github.com/mickamy/go-event-sourcing

# Optionally install a backend
go get github.com/mickamy/go-event-sourcing/stores/pgx
```

## Example

```go
import (
    "context"
    
	"github.com/jackc/pgx/v5/pgxpool"
    
	"github.com/mickamy/go-event-sourcing"
    "github.com/mickamy/go-event-sourcing/stores/pgx"
)

func main() {
    ctx := context.Background()

    pool, _ := pgxpool.New(ctx, "postgres://postgres:password@localhost:5432/ges?sslmode=disable")
    store := pgx.NewEventStore(pool,
        pgxstore.WithTypeRegistry(map[string]ges.EventCodec{
            "AccountOpened":   AccountOpenedCodec{},
            "MoneyDeposited":  MoneyDepositedCodec{},
        }),
    )

    // Define your aggregate
    acc := &Account{}

    // Append new events
    _, err := store.Append(ctx, acc.StreamID(), acc.Version(), acc.PendingEvents(), nil)
    if err != nil {
        log.Fatal(err)
    }

    // Load and rehydrate
    events, last, _ := store.Load(ctx, acc.StreamID(), 0)
    acc.Restore(events)
    fmt.Printf("Restored balance=%d (version=%d)\n", acc.Balance, last)
}
```

## Interface Overview

```go
type EventStore interface {
    Load(ctx context.Context, streamID string, fromVersion int64) ([]Event, int64, error)
    Append(ctx context.Context, streamID string, expectedVersion int64, events []Event, md Metadata) (int64, error)
    SaveSnapshot(ctx context.Context, streamID string, version int64, state any) error
    LoadSnapshot(ctx context.Context, streamID string) (Snapshot, error)
}
```

## License

[MIT](./LICENSE)
