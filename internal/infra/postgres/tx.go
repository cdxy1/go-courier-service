package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type poolAdapter struct{ pool *pgxpool.Pool }
type txAdapter struct{ tx pgx.Tx }
type txContextKey struct{}

func NewPoolAdapter(p *pgxpool.Pool) DB { return &poolAdapter{pool: p} }

func NewTxAdapter(t pgx.Tx) DB { return &txAdapter{tx: t} }

func (p *poolAdapter) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := p.pool.Exec(ctx, sql, args...)
	return err
}
func (p *poolAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}
func (p *poolAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (t *txAdapter) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := t.tx.Exec(ctx, sql, args...)
	return err
}
func (t *txAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}
func (t *txAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

type TxManager struct{ pool *pgxpool.Pool }

func NewTxManager(pool *pgxpool.Pool) *TxManager { return &TxManager{pool: pool} }

func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	ctx = context.WithValue(ctx, txContextKey{}, NewTxAdapter(tx))
	if err := fn(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (m *TxManager) PoolDB() DB { return NewPoolAdapter(m.pool) }

func DBFromContext(ctx context.Context, pool *pgxpool.Pool) DB {
	if v := ctx.Value(txContextKey{}); v != nil {
		if db, ok := v.(DB); ok {
			return db
		}
	}
	return NewPoolAdapter(pool)
}
