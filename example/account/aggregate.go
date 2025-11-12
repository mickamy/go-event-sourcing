package main

import (
	"fmt"

	"github.com/mickamy/go-event-sourcing"
)

// Account is the aggregate root that enforces domain rules and emits events.
type Account struct {
	id      string
	owner   string
	balance int64
	version int64       // current version (after applying pending)
	pend    []ges.Event // uncommitted domain events
	opened  bool
}

func (a *Account) record(e ges.Event) {
	a.Apply(e)
	a.pend = append(a.pend, e)
}

func (a *Account) Balance() int64 {
	return a.balance
}

// Handle routes a command to domain logic and records resulting events.
func (a *Account) Handle(cmd any) error {
	switch c := cmd.(type) {
	case OpenAccountCommand:
		if a.opened {
			return fmt.Errorf("account already opened")
		}
		if c.AccountID == "" {
			return fmt.Errorf("empty account id")
		}
		if c.Initial < 0 {
			return fmt.Errorf("initial balance cannot be negative")
		}
		a.record(AccountOpened{AccountID: c.AccountID, Owner: c.Owner, Initial: c.Initial})
		return nil

	case DepositCommand:
		if !a.opened {
			return fmt.Errorf("account not opened")
		}
		if c.Amount <= 0 {
			return fmt.Errorf("invalid deposit amount")
		}
		a.record(MoneyDeposited{Amount: c.Amount})
		return nil
	}

	return fmt.Errorf("unknown command type %T", cmd)
}

func (a *Account) StreamID() string { return "Account:" + a.id }

func (a *Account) Apply(e ges.Event) {
	switch ev := e.(type) {
	case AccountOpened:
		a.id = ev.AccountID
		a.owner = ev.Owner
		a.balance = ev.Initial
		a.opened = true
	case MoneyDeposited:
		a.balance += ev.Amount
	}
	a.version++
}

func (a *Account) Restore(events []ges.Event) {
	for _, e := range events {
		a.Apply(e)
	}
}

func (a *Account) Flush() ([]ges.Event, int64) {
	n := int64(len(a.pend))
	expected := a.version - n
	evs := make([]ges.Event, len(a.pend))
	copy(evs, a.pend)
	a.pend = nil
	return evs, expected
}

func (a *Account) Version() int64 { return a.version }

var _ ges.Aggregate = (*Account)(nil)
