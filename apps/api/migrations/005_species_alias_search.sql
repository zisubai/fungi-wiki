CREATE UNIQUE INDEX IF NOT EXISTS idx_species_aliases_unique_name
    ON species_aliases (species_id, LOWER(alias_name));

CREATE INDEX IF NOT EXISTS idx_species_aliases_name_trgm
    ON species_aliases USING GIN (alias_name gin_trgm_ops);
