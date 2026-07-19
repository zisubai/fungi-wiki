CREATE TABLE IF NOT EXISTS import_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(80) NOT NULL DEFAULT 'species',
    source_filename VARCHAR(255) NOT NULL,
    total_rows INTEGER NOT NULL DEFAULT 0,
    success_rows INTEGER NOT NULL DEFAULT 0,
    failed_rows INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(40) NOT NULL DEFAULT 'processing',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT import_batches_status_check CHECK (status IN ('processing', 'completed', 'failed'))
);

CREATE TABLE IF NOT EXISTS import_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES import_batches(id) ON DELETE CASCADE,
    row_number INTEGER NOT NULL,
    raw_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    species_id UUID REFERENCES species(id),
    status VARCHAR(40) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT import_rows_status_check CHECK (status IN ('imported', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_import_rows_batch_id ON import_rows (batch_id);
CREATE INDEX IF NOT EXISTS idx_import_batches_created_at ON import_batches (created_at DESC);
