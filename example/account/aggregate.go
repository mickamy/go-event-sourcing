package main

import (
	"fmt"

	ges "github.com/mickamy/go-event-sourcing"
)

type Account struct {
	ges.Base
	owner   string
	balance int64
	opened  bool
}

func (a *Account) Balance() int64 { return a.balance }

// Handle routes a command and records resulting events via Raise.
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
		a.Raise(AccountOpened{AccountID: c.AccountID, Owner: c.Owner, Initial: c.Initial})
		return nil

	case DepositCommand:
		if !a.opened {
			return fmt.Errorf("account not opened")
		}
		if c.Amount <= 0 {
			return fmt.Errorf("invalid deposit amount")
		}
		a.Raise(MoneyDeposited{Amount: c.Amount})
		return nil
	}
	return fmt.Errorf("unknown command type %T", cmd)
}

// applier: state mutation per event (used by Base.Apply/Raise).
func (a *Account) when(e ges.Event) {
	switch ev := e.(type) {
	case AccountOpened:
		a.SetStreamID("Account:" + ev.AccountID)
		a.owner = ev.Owner
		a.balance = ev.Initial
		a.opened = true
	case MoneyDeposited:
		a.balance += ev.Amount
	}
}

// Restore replays committed events (helper used by repository).
func (a *Account) Restore(events []ges.Event) {
	for _, e := range events {
		a.Apply(e) // Base.Apply → when() → version++
	}
}

var _ ges.Aggregate = (*Account)(nil)
