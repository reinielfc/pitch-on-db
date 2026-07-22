-- +goose Up
CREATE VIEW pigeon_summaries AS
SELECT
    p.id,
    p.name,
    p.ring_number,
    p.sex,
    p.sex_confidence,
    p.status,
    p.acquired_date,
    p.acquired_via,
    p.created_at
FROM pigeons p;

-- +goose Down
DROP VIEW IF EXISTS pigeon_summaries;
