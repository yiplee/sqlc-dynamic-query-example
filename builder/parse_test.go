package builder

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		stmt    string
		table   string
		columns []string
	}{
		{
			name:    "lowercase simple",
			stmt:    "select a,b,c FROM test;",
			table:   "test",
			columns: []string{"a", "b", "c"},
		},
		{
			name: "mixed case multiline",
			stmt: `
			 SELECT
				id ,
				 name
				FROM
				 users_table
				;
			`,
			table:   "users_table",
			columns: []string{"id", "name"},
		},
		{
			name:    "no match",
			stmt:    "update table set a=1",
			table:   "",
			columns: nil,
		},
		{
			name:    "leading whitespace",
			stmt:    "  \n\tselect  col1 ,  col2  from  my_tab  ",
			table:   "my_tab",
			columns: []string{"col1", "col2"},
		},
		{
			name:    "quoted column names",
			stmt:    "SELECT `a`, \"b\", 'c' FROM t1;",
			table:   "t1",
			columns: []string{"a", "b", "c"},
		},
		{
			name: "sqlc generated query",
			stmt: `-- name: ListUsers :many
SELECT id, name, -- comment
email, age, created_at, updated_at FROM users
`,
			table:   "users",
			columns: []string{"id", "name", "email", "age", "created_at", "updated_at"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, cols := Parse(tt.stmt)

			if !reflect.DeepEqual(cols, tt.columns) {
				t.Fatalf("columns = %#v, want %#v", cols, tt.columns)
			}

			if table != tt.table {
				t.Fatalf("table = %q, want %q", table, tt.table)
			}
		})
	}
}
