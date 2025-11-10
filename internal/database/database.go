package database

func Connect(ctx context.Context, connectionUrl string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(ctx, connectionUrl)
}

func Migrate(ctx context.Context, dbPool *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id BIGINT PRIMARY KEY,
			balance NUMERIC(20, 2) NOT NULL CHECK (balance >= 0)
		);`,
		`CREATE TABLE IF NOT EXISTS transactions_log (
			id SERIAL PRIMARY KEY,
			source_account_id BIGINT NOT NULL,
			destination_account_id BIGINT NOT NULL,
			initial_source_balance NUMERIC(20, 2) NOT NULL,
			initial_destination_balance NUMERIC(20, 2) NOT NULL,
			amount NUMERIC(20, 2) NOT NULL CHECK (amount > 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			CONSTRAINT fk_source_account FOREIGN KEY (source_account_id) REFERENCES accounts(id),
			CONSTRAINT fk_destination_account FOREIGN KEY (destination_account_id) REFERENCES accounts(id)
		);`,
	}
	for _, stmt := range statements {
		if _, err := dbPool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func CreateAccount(ctx context.Context, dbPool *pgxpool.Pool, accountID int64, initialBalance decimal.Decimal) (*types.AccountResponse, error) {
	_, err := dbPool.Exec(ctx, "INSERT INTO accounts (id, balance) VALUES ($1, $2)", accountID, initialBalance)
	if err != nil {
		return nil, err
	}
	return &types.AccountResponse{
		AccountID: accountID,
		Balance:   initialBalance,
	}, nil
}

func GetAccountByID(ctx context.Context, dbPool *pgxpool.Pool, accountID int64) (*types.AccountResponse, error) {
	var balance decimal.Decimal
	err := dbPool.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&balance)
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

func CreateTransaction(ctx context.Context, dbPool *pgxpool.Pool, sourceAccountID, destinationAccountID int64, amount decimal.Decimal) error {
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var sourceBalanceStr, destinationBalanceStr decimal.Decimal
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
