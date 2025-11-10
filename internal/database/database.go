package database

func Connect(ctx context.Context, connectionUrl string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(ctx, connectionUrl)
}

func Migrate(ctx context.Context, dbPool *pgxpool.Pool) error {
	statements := []string{}
	for _, stmt := range statements {
		if _, err := dbPool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}