package builder

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

type selector struct {
	raw           DBTX
	expectedQuery string
	sb            sq.SelectBuilder
}

func (r *selector) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// Prepared statements are intentionally not intercepted:
	// the builder's placeholder args are not available at stmt exec time.
	return r.raw.PrepareContext(ctx, query)
}

func (r *selector) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// This wrapper is intended for SELECT interception; keep Exec passthrough.
	return r.raw.ExecContext(ctx, query, args...)
}

func (r *selector) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if r.expectedQuery != "" && query != r.expectedQuery {
		return r.raw.QueryContext(ctx, query, args...)
	}

	overriddenQuery, overriddenArgs, err := r.sb.ToSql()
	if err != nil {
		return nil, err
	}
	return r.raw.QueryContext(ctx, overriddenQuery, overriddenArgs...)
}

func (r *selector) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if r.expectedQuery != "" && query != r.expectedQuery {
		return r.raw.QueryRowContext(ctx, query, args...)
	}

	overriddenQuery, overriddenArgs, err := r.sb.ToSql()
	if err != nil {
		// Can't return an error from QueryRowContext; fall back to raw behavior.
		return r.raw.QueryRowContext(ctx, query, args...)
	}
	return r.raw.QueryRowContext(ctx, overriddenQuery, overriddenArgs...)
}
