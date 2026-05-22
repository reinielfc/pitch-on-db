-- name: ListPigeons :many
SELECT id, name, created_at FROM pigeons ORDER BY id;

-- name: GetPigeon :one
SELECT id, name, created_at FROM pigeons WHERE id = $1;

-- name: CreatePigeon :one
INSERT INTO pigeons (name) VALUES ($1) RETURNING id, name, created_at;
