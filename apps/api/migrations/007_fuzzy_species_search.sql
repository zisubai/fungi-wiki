CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_species_slug_trgm
ON species USING GIN (slug gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_species_summary_trgm
ON species USING GIN (summary gin_trgm_ops);
