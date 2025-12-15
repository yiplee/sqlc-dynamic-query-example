-- name: ListUsers :many
SELECT * FROM users;

-- name: ListPosts :many
SELECT * FROM posts;

-- name: ListUserPosts :many
SELECT * FROM user_posts;
