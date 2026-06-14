-- name: ListPigeons :many
SELECT *
FROM pigeons
ORDER BY id;

-- name: GetPigeon :one
SELECT *
FROM pigeons
WHERE id = $1;

-- name: PigeonExists :one
SELECT EXISTS (
    SELECT 1
    FROM pigeons
    WHERE id = $1
);

-- name: CreatePigeon :one
INSERT INTO pigeons (name, band_number, birth_date, sex)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePigeon :one
UPDATE pigeons
SET
    name        = COALESCE(sqlc.narg('name'), name),
    band_number = CASE WHEN sqlc.arg('set_band_number')::bool THEN sqlc.narg('band_number') ELSE band_number END,
    birth_date  = CASE WHEN sqlc.arg('set_birth_date')::bool THEN sqlc.narg('birth_date') ELSE birth_date END,
    sex         = CASE WHEN sqlc.arg('set_sex')::bool THEN sqlc.narg('sex') ELSE sex END
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeletePigeon :exec
DELETE FROM pigeons
WHERE id = $1;
