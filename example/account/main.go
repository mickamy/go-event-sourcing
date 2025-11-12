package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mickamy/go-event-sourcing"
	"github.com/mickamy/go-event-sourcing/stores/pgx"
)

func main() {
	ctx := context.Background()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:password@localhost:5432/ges?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	defer pool.Close()

	store := pgx.NewEventStore(pool,
		pgx.WithTypeRegistry(map[string]ges.EventCodec{
			"AccountOpened":  ges.JSONCodec[AccountOpened](),
			"MoneyDeposited": ges.JSONCodec[MoneyDeposited](),
		}),
	)

	svc := NewAccountService(store)
	id := uuid.NewString()

	var cmd any

	// 1) Open account
	cmd = OpenAccountCommand{
		AccountID: id,
		Owner:     "Taro",
		Initial:   1000,
	}
	if err := svc.Handle(ctx, cmd, ges.Metadata{"tenant_id": "t1", "user_id": "u1"}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account opened: %+v\n", cmd)
	fmt.Println()

	// 2) Deposit
	cmd = DepositCommand{
		AccountID: id,
		Amount:    500,
	}
	if err := svc.Handle(ctx, cmd, ges.Metadata{"tenant_id": "t1", "user_id": "u1"}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account deposited: %+v\n", cmd)
	fmt.Println()

	// 3) Load and show balance (rehydrate)
	acc, err := NewAccountRepository(store).Load(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Restored account %s: balance=%d (version=%d)\n", id, acc.Balance(), acc.Version())
}
