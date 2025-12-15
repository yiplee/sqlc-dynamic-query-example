package tutorial

import (
	"context"
	"database/sql"
	"sqlc-dynamic-query-example/builder"

	sq "github.com/Masterminds/squirrel"
)

var (
	listUsersBuilder = builder.SelectBuilderFromStmt(listUsers).
				Where(sq.Eq{"deleted_at": nil}). // filter out deleted users
				Limit(500)                       // limit the number of users to 500 default to avoid overwhelming the database
	listUserPostsBuilder = builder.SelectBuilderFromStmt(listUserPosts).
				Where(sq.Eq{"deleted_at": nil}). // filter out deleted user posts
				Limit(500)                       // limit the number of user posts to 500 default to avoid overwhelming the database
)

func ListUsers(ctx context.Context, db DBTX, fn func(sb sq.SelectBuilder) sq.SelectBuilder) ([]User, error) {
	sb := listUsersBuilder
	if fn != nil {
		sb = fn(sb)
	}
	db = builder.Select(db, sb)
	q := New(db)
	return q.ListUsers(ctx)
}

func FindUser(ctx context.Context, db DBTX, fn func(sb sq.SelectBuilder) sq.SelectBuilder) (User, error) {
	users, err := ListUsers(ctx, db, func(sb sq.SelectBuilder) sq.SelectBuilder {
		if fn != nil {
			sb = fn(sb)
		}
		// add limit 1 to the query to avoid returning multiple users
		return sb.Limit(1)
	})
	if err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, sql.ErrNoRows
	}
	return users[0], nil
}

func ListUserPosts(ctx context.Context, db DBTX, fn func(sb sq.SelectBuilder) sq.SelectBuilder) ([]UserPost, error) {
	sb := listUserPostsBuilder
	if fn != nil {
		sb = fn(sb)
	}
	db = builder.Select(db, sb)
	q := New(db)
	return q.ListUserPosts(ctx)
}

func FindUserPost(ctx context.Context, db DBTX, fn func(sb sq.SelectBuilder) sq.SelectBuilder) (UserPost, error) {
	userPosts, err := ListUserPosts(ctx, db, func(sb sq.SelectBuilder) sq.SelectBuilder {
		if fn != nil {
			sb = fn(sb)
		}
		// add limit 1 to the query to avoid returning multiple user posts
		return sb.Limit(1)
	})
	if err != nil {
		return UserPost{}, err
	}

	if len(userPosts) == 0 {
		return UserPost{}, sql.ErrNoRows
	}
	return userPosts[0], nil
}
