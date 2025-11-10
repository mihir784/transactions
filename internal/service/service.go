package service

import (
	"encoding/json"
	"net/http"
)

type Service struct {
	dbPool *database.dbPool
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

	account, err := s.dbPool.CreateAccount(ctx, req.AccountID, initialBalance)
	if err != nil {
		if s.dbPool.IsUniqueViolationError(err) {
			writeJsonResponse(w, http.StatusConflict, types.ErrorResponse{Error: "account with the given account_id already exists"})
			return
		}
		writeJsonResponse(w, http.StatusInternalServerError, types.ErrorResponse{Error: "failed to create account"})
		return
	}

	writeJsonResponse(w, http.StatusNoContent)
}

func (s *Service) getAccountByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.split(strings.TrimPrefix(r.URL.Path, "/accounts/"), "/")
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
	account, err := s.dbPool.GetAccountByID(ctx, accountID)
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
	req := types.CreateTransactionRequest
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

	err := s.dbPool.CreateTransaction(ctx, req.SourceAccountID, req.DestinationAccountID, req.Amount)
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


	writeJsonResponse(w, http.StatusNoContent)
}
