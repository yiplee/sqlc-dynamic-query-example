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

func Select(raw DBTX, sb sq.SelectBuilder) DBTX {
	return &selecter{
		raw: raw,
		sb:  sb,
	}
}
