package pgx_test

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mickamy/go-event-sourcing"
	"github.com/mickamy/go-event-sourcing/internal/storetest"
	"github.com/mickamy/go-event-sourcing/stores/pgx"
)

func TestStore_Compliance(t *testing.T) {
	t.Parallel()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:password@localhost:5432/ges?sslmode=disable"
	}

	ctx := t.Context()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	storetest.Run(t, func(t *testing.T) ges.EventStore {
		t.Helper()
		return pgx.NewEventStore(
			pool,
			pgx.WithTypeRegistry(storetest.Registry()),
		)
	})
}
