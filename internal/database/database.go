package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

)

func Connect(ctx context.Context, connectionUrl string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, connectionUrl)
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
