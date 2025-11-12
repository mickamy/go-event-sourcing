package main

// AccountOpened is emitted when a new account is created.
type AccountOpened struct {
	AccountID string
	Owner     string
	Initial   int64
}

func (AccountOpened) EventType() string { return "AccountOpened" }

// MoneyDeposited is emitted when funds are deposited to an account.
type MoneyDeposited struct {
	Amount int64
}

func (MoneyDeposited) EventType() string { return "MoneyDeposited" }
