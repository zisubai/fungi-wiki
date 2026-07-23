CREATE TABLE species_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    change_type VARCHAR(40) NOT NULL,
    source_table VARCHAR(80) NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (species_id, version_number)
);

CREATE INDEX idx_species_versions_species_created
    ON species_versions (species_id, created_at DESC);

CREATE OR REPLACE FUNCTION capture_species_version(
    target_species_id UUID,
    event_type TEXT,
    event_source TEXT
) RETURNS VOID AS $$
DECLARE
    next_version INTEGER;
    species_snapshot JSONB;
BEGIN
    SELECT COALESCE(MAX(version_number), 0) + 1
      INTO next_version
      FROM species_versions
     WHERE species_id = target_species_id;

    SELECT jsonb_build_object(
        'species', to_jsonb(s),
        'aliases', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM species_aliases x WHERE x.species_id = s.id), '[]'::jsonb),
        'functions', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM species_functions x WHERE x.species_id = s.id), '[]'::jsonb),
        'cultureConditions', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM culture_conditions x WHERE x.species_id = s.id), '[]'::jsonb),
        'evidences', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM evidences x WHERE x.species_id = s.id), '[]'::jsonb),
        'applicationCases', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM application_cases x WHERE x.species_id = s.id), '[]'::jsonb)
    ) INTO species_snapshot
    FROM species s
    WHERE s.id = target_species_id;

    IF species_snapshot IS NOT NULL THEN
        INSERT INTO species_versions(species_id, version_number, change_type, source_table, snapshot)
        VALUES(target_species_id, next_version, event_type, event_source, species_snapshot);
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION capture_species_row_version()
RETURNS TRIGGER AS $$
DECLARE
    target_id UUID;
BEGIN
    target_id := CASE WHEN TG_OP = 'DELETE' THEN OLD.species_id ELSE NEW.species_id END;
    PERFORM capture_species_version(target_id, lower(TG_OP), TG_TABLE_NAME);
    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION capture_species_master_version()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM capture_species_version(NEW.id, lower(TG_OP), TG_TABLE_NAME);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_species_version
AFTER INSERT OR UPDATE ON species
FOR EACH ROW EXECUTE FUNCTION capture_species_master_version();

CREATE TRIGGER trg_species_aliases_version
AFTER INSERT OR UPDATE OR DELETE ON species_aliases
FOR EACH ROW EXECUTE FUNCTION capture_species_row_version();

CREATE TRIGGER trg_species_functions_version
AFTER INSERT OR UPDATE OR DELETE ON species_functions
FOR EACH ROW EXECUTE FUNCTION capture_species_row_version();

CREATE TRIGGER trg_culture_conditions_version
AFTER INSERT OR UPDATE OR DELETE ON culture_conditions
FOR EACH ROW EXECUTE FUNCTION capture_species_row_version();

CREATE TRIGGER trg_evidences_version
AFTER INSERT OR UPDATE OR DELETE ON evidences
FOR EACH ROW EXECUTE FUNCTION capture_species_row_version();

CREATE TRIGGER trg_application_cases_version
AFTER INSERT OR UPDATE OR DELETE ON application_cases
FOR EACH ROW EXECUTE FUNCTION capture_species_row_version();

INSERT INTO species_versions(species_id, version_number, change_type, source_table, snapshot)
SELECT s.id, 1, 'baseline', 'migration', jsonb_build_object(
    'species', to_jsonb(s),
    'aliases', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM species_aliases x WHERE x.species_id = s.id), '[]'::jsonb),
    'functions', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM species_functions x WHERE x.species_id = s.id), '[]'::jsonb),
    'cultureConditions', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM culture_conditions x WHERE x.species_id = s.id), '[]'::jsonb),
    'evidences', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM evidences x WHERE x.species_id = s.id), '[]'::jsonb),
    'applicationCases', COALESCE((SELECT jsonb_agg(to_jsonb(x) ORDER BY x.created_at) FROM application_cases x WHERE x.species_id = s.id), '[]'::jsonb)
) FROM species s;
