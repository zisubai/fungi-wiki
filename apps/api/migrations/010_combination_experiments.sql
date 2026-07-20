CREATE TABLE IF NOT EXISTS combination_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    combination_record_id UUID NOT NULL REFERENCES combination_recommendation_records(id) ON DELETE CASCADE,
    outcome VARCHAR(40) NOT NULL CHECK (outcome IN ('compatible', 'incompatible', 'inconclusive')),
    temperature NUMERIC,
    ph NUMERIC CHECK (ph IS NULL OR (ph >= 0 AND ph <= 14)),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_combination_experiments_record_created
    ON combination_experiments(combination_record_id, created_at DESC);
