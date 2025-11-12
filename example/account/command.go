package main

// OpenAccountCommand represents an intent to create a new account.
type OpenAccountCommand struct {
	AccountID string
	Owner     string
	Initial   int64
}

// DepositCommand represents an intent to increase the account balance.
type DepositCommand struct {
	AccountID string
	Amount    int64
}
