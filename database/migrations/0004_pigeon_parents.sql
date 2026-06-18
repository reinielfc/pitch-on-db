-- +goose Up
ALTER TABLE pigeons
    ADD COLUMN father_id BIGINT REFERENCES pigeons(id) ON DELETE SET NULL,
    ADD COLUMN mother_id BIGINT REFERENCES pigeons(id) ON DELETE SET NULL;

ALTER TABLE pigeons
    ADD CONSTRAINT pigeons_father_not_self CHECK (father_id IS DISTINCT FROM id),
    ADD CONSTRAINT pigeons_mother_not_self CHECK (mother_id IS DISTINCT FROM id);

CREATE INDEX idx_pigeons_father_id ON pigeons (father_id);
CREATE INDEX idx_pigeons_mother_id ON pigeons (mother_id);

-- +goose Down
DROP INDEX IF EXISTS idx_pigeons_mother_id;
DROP INDEX IF EXISTS idx_pigeons_father_id;

ALTER TABLE pigeons
    DROP CONSTRAINT IF EXISTS pigeons_mother_not_self,
    DROP CONSTRAINT IF EXISTS pigeons_father_not_self,
    DROP COLUMN mother_id,
    DROP COLUMN father_id;
