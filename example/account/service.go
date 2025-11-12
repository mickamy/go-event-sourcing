package main

import (
	"context"

	"github.com/mickamy/go-event-sourcing"
)

// AccountService orchestrates command handling using repository + store.
type AccountService struct {
	repo  *AccountRepository
	store ges.EventStore
}

// NewAccountService wires a repository and store together.
func NewAccountService(store ges.EventStore) *AccountService {
	return &AccountService{
		repo:  NewAccountRepository(store),
		store: store,
	}
}

// Handle executes a command end-to-end: load → Handle → append.
func (s *AccountService) Handle(ctx context.Context, cmd any, md ges.Metadata) error {
	// Determine target aggregate ID from the command.
	id := extractAccountID(cmd)
	acc, err := s.repo.Load(ctx, id)
	if err != nil {
		return err
	}

	// Route to domain logic.
	if err := acc.Handle(cmd); err != nil {
		return err
	}

	// Persist resulting events.
	return s.repo.Save(ctx, acc, md)
}

// extractAccountID is a tiny helper for this sample.
// In a real app, consider a command interface exposing AggregateID().
func extractAccountID(cmd any) string {
	switch c := cmd.(type) {
	case OpenAccountCommand:
		return c.AccountID
	case DepositCommand:
		return c.AccountID
	default:
		return ""
	}
}
