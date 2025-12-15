package builder

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

type selecter struct {
	raw DBTX
	sb  sq.SelectBuilder
}

func (r *selecter) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query, _ = r.sb.MustSql()
	return r.raw.PrepareContext(ctx, query)
}

func (r *selecter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query, args = r.sb.MustSql()
	return r.raw.ExecContext(ctx, query, args...)
}

func (r *selecter) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query, args = r.sb.MustSql()
	return r.raw.QueryContext(ctx, query, args...)
}

func (r *selecter) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query, args = r.sb.MustSql()
	return r.raw.QueryRowContext(ctx, query, args...)
}
