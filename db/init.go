package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func InitDB(ctx context.Context) (*Queries, error) {
	// TODO: `connStr` should be retrieved from elsewhere NOT hardcoded
	connStr := "user=test dbname=test password=password123 sslmode=disable host=localhost port=5432"
	pcfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		panic(err)
	}
	d := stdlib.OpenDB(*pcfg)
	defer d.Close()

	return New(d), nil
}
