package types

import "github.com/shopspring/decimal"

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateAccountRequest struct {
	AccountID 		int64 	`json:"account_id"`
	InitialBalance	string	`json:"initial_balance"`
}

type AccountResponse struct {
	AccountID	int64 			`json:"account_id"`
	Balance		decimal.Decimal	`json:"balance"`
}

type CreateTransactionRequest struct {
	SourceAccountID			int64			`json:"source_account_id"`
	DestinationAccountID	int64			`json:"destination_account_id"`
	Amount					decimal.Decimal	`json:"amount"`
}