-- name: FindPigeonByID :one
SELECT * FROM pigeons WHERE id = $1;

-- name: UpsertPigeon :one
INSERT INTO pigeons (id, name, ring_number, sex, sex_confidence, status, acquired_date, acquired_via, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    ring_number = EXCLUDED.ring_number,
    sex = EXCLUDED.sex,
    sex_confidence = EXCLUDED.sex_confidence,
    status = EXCLUDED.status,
    acquired_date = EXCLUDED.acquired_date,
    acquired_via = EXCLUDED.acquired_via,
    updated_at = now()
RETURNING *;

-- name: ListPigeons :many
SELECT *
FROM pigeon_summaries
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPigeons :one
SELECT COUNT(*) FROM pigeons;
