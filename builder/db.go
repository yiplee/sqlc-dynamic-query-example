package builder

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// SelectForQuery wraps DBTX and overrides only the matching sqlc query.
// If expectedQuery is empty, all QueryContext/QueryRowContext calls are overridden.
func SelectForQuery(raw DBTX, expectedQuery string, sb sq.SelectBuilder) DBTX {
	return &selector{
		raw:           raw,
		expectedQuery: expectedQuery,
		sb:            sb,
	}
}

func Select(raw DBTX, sb sq.SelectBuilder) DBTX {
	// Backwards-compatible behavior: override all query calls.
	return SelectForQuery(raw, "", sb)
}
