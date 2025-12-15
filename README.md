## Introduction

When building database-driven applications in Go, developers often face a dilemma: choose type safety with [sqlc](https://sqlc.dev/) or flexibility with query builders like [Squirrel](https://github.com/Masterminds/squirrel). What if we could have both?

This article demonstrates a technique to combine the best of both worlds—leveraging sqlc's type-safe code generation while enabling dynamic query building with Squirrel. We'll explore how to intercept database operations and replace static SQL queries with dynamically generated ones, all while preserving sqlc's type-safe models and scanning logic.

## The Problem

SQLC is excellent for generating type-safe Go code from SQL queries. It provides compile-time safety, eliminates runtime SQL errors, and generates efficient code. However, it requires all queries to be defined statically at code generation time.

On the other hand, query builders like Squirrel offer flexibility to construct queries dynamically based on runtime conditions—perfect for building search filters, pagination, or conditional WHERE clauses. But they lack the type safety and code generation benefits of sqlc.

The challenge: **How can we use sqlc's generated models and scanning logic while still building queries dynamically?**

## The Solution

The key insight is to intercept the database execution layer. Instead of modifying sqlc's generated code, we wrap the database connection with a custom implementation that replaces static queries with dynamically generated ones at runtime.

Here's how it works:

1. **Parse SQL statements** to extract table names and column lists from sqlc-generated query constants
2. **Build dynamic queries** using Squirrel based on the parsed metadata
3. **Intercept execution** by wrapping the database connection with a custom `DBTX` implementation
4. **Preserve type safety** by keeping sqlc's generated models and scanning logic intact

## Prerequisites

Before we dive in, make sure you have:

- Go 1.21 or later
- [sqlc installed](https://docs.sqlc.dev/en/latest/overview/install.html)
- A MySQL database (or modify the examples for your preferred database)

## Configuration

### Critical: Disable Prepared Statements

**Important**: You must set `emit_prepared_queries: false` in your `sqlc.yaml` configuration. This is crucial because prepared statements cache the SQL query, preventing us from replacing it with a dynamically generated one.

```yaml
version: "2"
sql:
    - engine: "mysql"
      queries: "sql/query.sql"
      schema: "sql/schema.sql"
      gen:
          go:
              package: "tutorial"
              out: "tutorial"
              # emit_prepared_queries must be false for the queries will be overridden by the dynamic queries
              emit_prepared_queries: false
```

### SQL Query Requirements

For any table that needs dynamic query support, you must include a `SELECT * FROM table` query in your `query.sql` file. This query serves two purposes:

1. **Code generation**: It tells sqlc to generate the model and query function
2. **Metadata extraction**: The query string is parsed to extract the table name and column list

Example:

```sql
-- name: ListUsers :many
SELECT * FROM users;

-- name: ListPosts :many
SELECT * FROM posts;
```

The `SELECT *` is parsed by our `builder.Parse()` function to extract:
- Table name: `users`, `posts`, etc.
- Column list: All columns from the table

**Note**: The actual SQL executed will be replaced by Squirrel, so the `SELECT *` is only used for metadata extraction during initialization.

## Architecture Overview

Let's examine the project structure to understand how everything fits together:

```
.
├── builder/           # Core dynamic query building logic
│   ├── db.go         # DBTX interface and Select wrapper
│   ├── parse.go      # SQL statement parsing (extracts table/columns)
│   └── select.go     # SelectBuilder interceptor implementation
├── tutorial/         # sqlc-generated code
│   ├── db.go         # Generated DBTX interface and Queries struct
│   ├── models.go     # Generated models (User, Post, etc.)
│   ├── query.sql.go  # Generated query functions
│   └── builder.go    # Dynamic query wrapper functions
├── sql/
│   ├── schema.sql    # Database schema
│   └── query.sql     # SQL queries for sqlc
├── example.go        # Usage example
└── sqlc.yaml         # sqlc configuration
```

## Implementation Deep Dive

### 1. SQL Parsing

The foundation of our solution is the `Parse()` function in `builder/parse.go`. It extracts table names and columns from SQL SELECT statements:

```go
func Parse(stmt string) (table string, columns []string)
```

This function handles:
- SQL comments (both `--` and `/* */`)
- Quoted identifiers (backticks, double quotes, single quotes)
- Multi-line statements
- Case-insensitive matching

The parser strips comments, extracts the SELECT clause and FROM clause, then splits columns by commas while respecting quoted identifiers.

### 2. SelectBuilder Creation

Once we've parsed the SQL statement, we convert it into a Squirrel SelectBuilder:

```go
func SelectBuilderFromStmt(stmt string) sq.SelectBuilder {
    table, columns := Parse(stmt)
    return sq.Select(columns...).From(table)
}
```

This creates a base SelectBuilder that we can then customize with additional conditions, ordering, and limits.

### 3. Query Interception

The magic happens in `builder/select.go`. The `selecter` struct implements the `DBTX` interface and intercepts all database operations:

```go
type selecter struct {
    raw DBTX
    sb  sq.SelectBuilder
}

func (r *selecter) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    query, args = r.sb.MustSql()  // Replace with dynamic query
    return r.raw.QueryContext(ctx, query, args...)
}
```

When sqlc's generated code calls `QueryContext` with the original static query, our interceptor replaces it with the dynamically built query from Squirrel. The same pattern applies to `QueryRowContext`, `ExecContext`, and `PrepareContext`.

### 4. Wrapper Functions

The high-level API in `tutorial/builder.go` combines everything together:

```go
func ListUsers(ctx context.Context, db DBTX, fn func(sb sq.SelectBuilder) sq.SelectBuilder) ([]User, error) {
    sb := listUsersBuilder
    if fn != nil {
        sb = fn(sb)
    }
    db = builder.Select(db, sb)
    q := New(db)
    return q.ListUsers(ctx)
}
```

These wrapper functions:
- Start with a base SelectBuilder (parsed from the sqlc query)
- Apply default filters (e.g., filtering out soft-deleted records)
- Allow callers to customize the query via the function parameter
- Wrap the database connection to intercept queries
- Call the original sqlc-generated function for type-safe scanning

## Usage Examples

Now let's see how to use this in practice:

### Basic Dynamic Query

```go
package main

import (
    "context"
    "database/sql"
    "sqlc-dynamic-query-example/tutorial"
    
    sq "github.com/Masterminds/squirrel"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    db, _ := sql.Open("mysql", "root:password@tcp(localhost:3306)/test")
    defer db.Close()
    
    ctx := context.Background()
    
    // Dynamic query with filters and ordering
    users, err := tutorial.ListUsers(ctx, db, func(sb sq.SelectBuilder) sq.SelectBuilder {
        return sb.
            Where(sq.GtOrEq{"age": 18}).
            OrderBy("created_at DESC").
            Limit(10)
    })
}
```

### Finding a Single Record

```go
// Find a single user with custom conditions
user, err := tutorial.FindUser(ctx, db, func(sb sq.SelectBuilder) sq.SelectBuilder {
    return sb.Where(sq.Eq{"email": "user@example.com"})
})
```

### Complex Filtering

```go
// Build complex queries based on runtime conditions
users, err := tutorial.ListUsers(ctx, db, func(sb sq.SelectBuilder) sq.SelectBuilder {
    conditions := sq.And{}
    
    if minAge > 0 {
        conditions = append(conditions, sq.GtOrEq{"age": minAge})
    }
    
    if searchTerm != "" {
        conditions = append(conditions, sq.Or{
            sq.Like{"name": "%" + searchTerm + "%"},
            sq.Like{"email": "%" + searchTerm + "%"},
        })
    }
    
    if len(conditions) > 0 {
        sb = sb.Where(conditions)
    }
    
    return sb.OrderBy("created_at DESC").Limit(limit)
})
```

## Benefits

This approach provides several advantages:

1. **Type Safety**: Leverage sqlc's generated models and scanning logic—no manual struct mapping
2. **Dynamic Queries**: Build queries dynamically with Squirrel's fluent API based on runtime conditions
3. **Best of Both Worlds**: Combine the safety of sqlc with the flexibility of query builders
4. **No Code Generation Changes**: Works with standard sqlc output—no need to modify generated code
5. **Backward Compatible**: Can still use static queries when dynamic behavior isn't needed

## Limitations and Considerations

While this technique is powerful, there are some limitations to be aware of:

1. **Prepared Statements**: Must disable `emit_prepared_queries` as mentioned above
2. **SELECT Only**: Currently supports SELECT queries (UPDATE support can be added similarly)
3. **Query Parsing**: Requires `SELECT * FROM table` format for metadata extraction
4. **Single Table**: Each query should target a single table (joins require special handling)

## Conclusion

By intercepting the database execution layer, we've successfully combined sqlc's type safety with Squirrel's dynamic query building capabilities. This approach allows you to:

- Maintain type safety with sqlc's generated code
- Build queries dynamically based on runtime conditions
- Avoid manual struct mapping and scanning
- Keep your codebase maintainable and type-safe

The technique is particularly useful for building APIs with dynamic filtering, search, and pagination while maintaining the safety guarantees that sqlc provides.

---

*This article demonstrates a practical technique for combining type-safe SQL code generation with dynamic query building. Feel free to adapt this approach to your own projects and requirements.*
