CREATE INDEX IF NOT EXISTS idx_species_safety_level ON species (safety_level);
CREATE INDEX IF NOT EXISTS idx_culture_conditions_temperature ON culture_conditions (temperature_min, temperature_max);
CREATE INDEX IF NOT EXISTS idx_culture_conditions_ph ON culture_conditions (ph_min, ph_max);
CREATE INDEX IF NOT EXISTS idx_species_functions_species_tag ON species_functions (species_id, function_tag_id);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_species_source_environment_trgm ON species USING GIN (source_environment gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_species_latin_name_trgm ON species USING GIN (latin_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_species_chinese_name_trgm ON species USING GIN (chinese_name gin_trgm_ops);
