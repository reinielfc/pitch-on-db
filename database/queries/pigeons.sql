-- name: CreatePigeon :one
INSERT INTO pigeons (name, band_number, birth_date, sex)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPigeons :many
SELECT * FROM pigeons ORDER BY id;

-- name: GetPigeon :one
SELECT * FROM pigeons WHERE id = $1;

-- name: GetPigeonSex :one
SELECT sex FROM pigeons WHERE id = $1;

-- name: CheckPigeonExists :one
SELECT EXISTS (
    SELECT 1
    FROM pigeons
    WHERE id = $1
);

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
DELETE FROM pigeons WHERE id = $1;

-- name: SetPigeonFather :exec
UPDATE pigeons
SET father_id = $1
WHERE id = $2;

-- name: SetPigeonMother :exec
UPDATE pigeons
SET mother_id = $1
WHERE id = $2;

-- name: GetPigeonFather :one
SELECT f.* FROM pigeons p
INNER JOIN pigeons f ON f.id = p.father_id
WHERE p.id = $1;

-- name: GetPigeonMother :one
SELECT m.* FROM pigeons p
INNER JOIN pigeons m ON m.id = p.mother_id
WHERE p.id = $1;

-- name: GetPigeonChildrenAsFather :many
SELECT c.* FROM pigeons p
INNER JOIN pigeons c ON c.father_id = p.id
WHERE p.id = $1;

-- name: GetPigeonChildrenAsMother :many
SELECT c.* FROM pigeons p
INNER JOIN pigeons c ON c.mother_id = p.id
WHERE p.id = $1;

-- name: CheckPigeonHasChildrenAsFather :one
SELECT EXISTS (
    SELECT 1
    FROM pigeons
    WHERE father_id = $1
);

-- name: CheckPigeonHasChildrenAsMother :one
SELECT EXISTS (
    SELECT 1
    FROM pigeons
    WHERE mother_id = $1
);
