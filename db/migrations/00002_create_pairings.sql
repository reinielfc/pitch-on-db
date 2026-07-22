-- +goose Up
CREATE TABLE pairings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_a_id    UUID NOT NULL REFERENCES pigeons (id) ON DELETE RESTRICT,
    partner_a_role  TEXT CHECK (partner_a_role IN ('sire', 'dam')),
    partner_b_id    UUID REFERENCES pigeons (id) ON DELETE RESTRICT,
    partner_b_role  TEXT CHECK (partner_b_role IN ('sire', 'dam')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_partners_distinct CHECK (
        partner_b_id IS NULL OR partner_a_id <> partner_b_id
    ),
    CONSTRAINT chk_roles_distinct CHECK (
        partner_a_role IS NULL OR partner_b_role IS NULL OR partner_a_role <> partner_b_role
    )
);

CREATE INDEX idx_pairings_partner_a ON pairings (partner_a_id);
CREATE INDEX idx_pairings_partner_b ON pairings (partner_b_id);

ALTER TABLE pigeons
    ADD COLUMN pairing_id UUID REFERENCES pairings (id) ON DELETE SET NULL;

CREATE INDEX idx_pigeons_pairing_id ON pigeons (pairing_id);

-- +goose Down
ALTER TABLE pigeons DROP COLUMN IF EXISTS pairing_id;
DROP TABLE IF EXISTS pairings;
