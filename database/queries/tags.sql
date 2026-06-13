-- name: GetPigeonTags :many
SELECT t.id, t.name
FROM tags t
JOIN pigeon_tags pt ON pt.tag_id = t.id
WHERE pt.pigeon_id = $1
ORDER BY t.name;

-- name: ClearPigeonTags :exec
DELETE FROM pigeon_tags
WHERE pigeon_id = $1;

-- name: UpsertTag :one
INSERT INTO tags (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;

-- name: AddPigeonTag :exec
INSERT INTO pigeon_tags (pigeon_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePigeonTag :exec
DELETE FROM pigeon_tags
WHERE pigeon_id = $1
  AND tag_id = (SELECT id FROM tags WHERE name = $2);
