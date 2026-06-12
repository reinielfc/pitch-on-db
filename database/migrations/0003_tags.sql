-- +goose Up
CREATE TABLE tags (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE pigeon_tags (
    pigeon_id BIGINT NOT NULL REFERENCES pigeons(id) ON DELETE CASCADE,
    tag_id    BIGINT NOT NULL REFERENCES tags(id)    ON DELETE CASCADE,
    PRIMARY KEY (pigeon_id, tag_id)
);

-- +goose Down
DROP TABLE IF EXISTS pigeon_tags;
DROP TABLE IF EXISTS tags;
