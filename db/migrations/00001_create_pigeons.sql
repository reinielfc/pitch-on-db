-- +goose Up
CREATE TABLE pigeons (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT,
    ring_number     TEXT,
    sex             TEXT NOT NULL DEFAULT 'unknown'
                    CHECK (sex IN ('male', 'female', 'unknown')),
    sex_confidence  TEXT
                    CHECK (sex_confidence IN ('confirmed', 'presumed')),
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'deceased', 'sold', 'lost')),
    acquired_date   DATE,
    acquired_via    TEXT NOT NULL DEFAULT 'unknown'
                    CHECK (acquired_via IN ('bred', 'purchased', 'gifted', 'captured', 'unknown')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ring_number is often unknown/unset for newly-imported birds, so it's nullable;
-- uniqueness only enforced once a value is actually present
CREATE UNIQUE INDEX idx_pigeons_ring_number ON pigeons (ring_number) WHERE ring_number IS NOT NULL;
CREATE INDEX idx_pigeons_status ON pigeons (status);

-- +goose Down
DROP TABLE IF EXISTS pigeons;
