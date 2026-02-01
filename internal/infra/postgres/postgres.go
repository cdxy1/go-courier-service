package postgres

import (
	"context"

	"github.com/cdxy1/go-courier-service/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConnection(cfg *config.Сonfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.Postgres.GetURL())

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
