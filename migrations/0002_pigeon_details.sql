-- +goose Up
ALTER TABLE pigeons
    ADD COLUMN band_number TEXT UNIQUE,
    ADD COLUMN birth_date  DATE,
    ADD COLUMN sex         TEXT CHECK (sex IN ('M', 'F'));

-- +goose Down
ALTER TABLE pigeons
    DROP COLUMN sex,
    DROP COLUMN birth_date,
    DROP COLUMN band_number;
