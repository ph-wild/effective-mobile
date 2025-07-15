package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func ConnectDB(ctx context.Context, connString string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", connString)
	if err != nil {
		return nil, errors.Wrap(err, "can't connect to DB")
	}
	return db, nil
}
