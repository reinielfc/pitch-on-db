-- +goose Up
UPDATE pigeons
SET properties = jsonb_set(
    COALESCE(properties, '{}'::jsonb),
    '{band_number}',
    to_jsonb(band_number),
    true
)
WHERE band_number IS NOT NULL;

ALTER TABLE pigeons
    DROP COLUMN band_number;

-- +goose Down
ALTER TABLE pigeons
    ADD COLUMN band_number TEXT UNIQUE;

UPDATE pigeons
SET band_number = properties ->> 'band_number'
WHERE properties ? 'band_number';

UPDATE pigeons
SET properties = properties - 'band_number'
WHERE properties ? 'band_number';