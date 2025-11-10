package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"transactions/internal/types"

)

type Service struct {
	DbPool *pgxpool.Pool
}

func writeJsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (s *Service) HandleRoot(w http.ResponseWriter, r *http.Request) {
	writeJsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Service) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.createAccount(w, r)
}

func (s *Service) HandleAccountByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.getAccountByID(w, r)
}

func (s *Service) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.createTransaction(w, r)
}

func (s *Service) createAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req types.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.AccountID <= 0 {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "account_id must be a positive integer"})
		return
	}

	if strings.TrimSpace(req.InitialBalance) == "" {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "initial_balance is required"})
		return
	}
	initialBalance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil || initialBalance.IsNegative() {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "initial_balance must be a non-negative decimal"})
		return
	}

	_, err = s.CreateAccountInDb(ctx, req.AccountID, initialBalance)
	if err != nil {
		if s.IsUniqueViolationError(err) {
			writeJsonResponse(w, http.StatusConflict, types.ErrorResponse{Error: "account with the given account_id already exists"})
			return
		}
		writeJsonResponse(w, http.StatusInternalServerError, types.ErrorResponse{Error: "failed to create account"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) getAccountByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/accounts/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "account_id is required"})
		return
	}
	accountIDStr := parts[0]
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil || accountID <= 0 {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "invalid account_id"})
		return
	}
	account, err := s.GetAccountByID(ctx, accountID)
	if err != nil {
		writeJsonResponse(w, http.StatusInternalServerError, types.ErrorResponse{Error: "failed to retrieve account"})
		return
	}
	if account == nil {
		writeJsonResponse(w, http.StatusNotFound, types.ErrorResponse{Error: "account not found"})
		return
	}
	writeJsonResponse(w, http.StatusOK, account)
}

func (s *Service) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req types.CreateTransactionRequest
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.SourceAccountID <= 0 {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "source_account_id must be a positive integer"})
		return
	}
	if req.DestinationAccountID <= 0 {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "destination_account_id must be a positive integer"})
		return
	}
	if req.SourceAccountID == req.DestinationAccountID {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "source_account_id and destination_account_id cannot be the same"})
		return
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "amount must be a positive decimal"})
		return
	}

	err := s.CreateTransaction(ctx, req.SourceAccountID, req.DestinationAccountID, req.Amount)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient") {
			writeJsonResponse(w, http.StatusBadRequest, types.ErrorResponse{Error: "insufficient funds in source account"})
			return
		}
		if strings.Contains(err.Error(), "notfound") {
			writeJsonResponse(w, http.StatusNotFound, types.ErrorResponse{Error: "one or both accounts not found"})
			return
		}
		writeJsonResponse(w, http.StatusInternalServerError, types.ErrorResponse{Error: "failed to create transaction"})
		return
	}


	w.WriteHeader(http.StatusNoContent)
}


func (s *Service) IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (s *Service) CreateAccountInDb(ctx context.Context, accountID int64, initialBalance decimal.Decimal) (*types.AccountResponse, error) {
	_, err := s.DbPool.Exec(ctx, "INSERT INTO accounts (id, balance) VALUES ($1, $2)", accountID, initialBalance)
	if err != nil {
		return nil, err
	}
	return &types.AccountResponse{
		AccountID: accountID,
		Balance:   initialBalance,
	}, nil
}

func (s *Service) GetAccountByID(ctx context.Context, accountID int64) (*types.AccountResponse, error) {
	var balance decimal.Decimal
	err := s.DbPool.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &types.AccountResponse{
		AccountID: accountID,
		Balance:   balance,
	}, nil
}

func (s *Service) CreateTransaction(ctx context.Context, sourceAccountID, destinationAccountID int64, amount decimal.Decimal) error {
	tx, err := s.DbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var sourceBalance, destinationBalance decimal.Decimal
	err = tx.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", sourceAccountID).Scan(&sourceBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("notfound:source_account")
		}
		return err
	}
	err = tx.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", destinationAccountID).Scan(&destinationBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("notfound:destination_account")
		}
		return err
	}
	if sourceBalance.LessThan(amount) {
		return errors.New("insufficient funds")
	}

	_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, sourceAccountID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, destinationAccountID)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, "INSERT INTO transactions_log (source_account_id, destination_account_id, initial_source_balance, initial_destination_balance, amount) VALUES ($1, $2, $3, $4, $5)", sourceAccountID, destinationAccountID, sourceBalance, destinationBalance, amount); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
