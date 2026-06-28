-- +goose Up
ALTER TABLE pigeons
	ADD COLUMN properties JSONB;

CREATE INDEX idx_pigeons_properties_gin ON pigeons USING GIN (properties);

-- +goose Down
DROP INDEX IF EXISTS idx_pigeons_properties_gin;

ALTER TABLE pigeons
	DROP COLUMN properties;
