package builder

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

type recordDB struct {
	lastQuery string
	lastArgs  []interface{}
}

func (r *recordDB) PrepareContext(context.Context, string) (*sql.Stmt, error) { return nil, nil }
func (r *recordDB) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (r *recordDB) QueryContext(_ context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	r.lastQuery = query
	r.lastArgs = append([]interface{}(nil), args...)
	return nil, nil
}
func (r *recordDB) QueryRowContext(context.Context, string, ...interface{}) *sql.Row { return nil }

func TestSelectForQuery_OnlyOverridesMatchingQuery(t *testing.T) {
	ctx := context.Background()
	raw := &recordDB{}

	sb := sq.Select("*").From("users").Where(sq.Eq{"id": 10})
	wrapped := SelectForQuery(raw, "expected", sb)

	// Non-matching query should pass through unchanged.
	if _, err := wrapped.QueryContext(ctx, "other", 1, 2, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw.lastQuery != "other" {
		t.Fatalf("query = %q, want %q", raw.lastQuery, "other")
	}
	if !reflect.DeepEqual(raw.lastArgs, []interface{}{1, 2, 3}) {
		t.Fatalf("args = %#v, want %#v", raw.lastArgs, []interface{}{1, 2, 3})
	}

	// Matching query should be overridden by the builder SQL.
	if _, err := wrapped.QueryContext(ctx, "expected"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantQuery, wantArgs, err := sb.ToSql()
	if err != nil {
		t.Fatalf("sb.ToSql() error: %v", err)
	}
	if raw.lastQuery != wantQuery {
		t.Fatalf("query = %q, want %q", raw.lastQuery, wantQuery)
	}
	if !reflect.DeepEqual(raw.lastArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", raw.lastArgs, wantArgs)
	}
}
