package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sqlc-dynamic-query-example/tutorial"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
)

var (
	driver = flag.String("driver", "mysql", "Database driver")
	dsn    = flag.String("dsn", "root:password@tcp(localhost:3306)/test", "MySQL connection string")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	// Open a connection to the database
	db, err := sql.Open(*driver, *dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	users, err := tutorial.ListUsers(ctx, db, func(sb sq.SelectBuilder) sq.SelectBuilder {
		return sb.Where(sq.GtOrEq{"age": 18}).
			OrderBy("created_at DESC").
			Limit(10)
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(users)
}
