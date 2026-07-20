CREATE OR REPLACE FUNCTION calculate_species_data_quality(target_species_id UUID)
RETURNS NUMERIC(5,2) AS $$
    SELECT (
        CASE WHEN NULLIF(BTRIM(s.latin_name), '') IS NOT NULL THEN 10 ELSE 0 END +
        CASE WHEN NULLIF(BTRIM(s.chinese_name), '') IS NOT NULL THEN 5 ELSE 0 END +
        CASE WHEN NULLIF(BTRIM(s.strain_number), '') IS NOT NULL THEN 5 ELSE 0 END +
        CASE WHEN NULLIF(BTRIM(s.source_environment), '') IS NOT NULL THEN 10 ELSE 0 END +
        CASE WHEN NULLIF(BTRIM(s.safety_level), '') IS NOT NULL THEN 10 ELSE 0 END +
        CASE WHEN NULLIF(BTRIM(s.summary), '') IS NOT NULL THEN 15 ELSE 0 END +
        CASE WHEN EXISTS (SELECT 1 FROM species_aliases sa WHERE sa.species_id = s.id) THEN 5 ELSE 0 END +
        CASE WHEN EXISTS (SELECT 1 FROM species_functions sf WHERE sf.species_id = s.id) THEN 15 ELSE 0 END +
        CASE WHEN EXISTS (SELECT 1 FROM culture_conditions cc WHERE cc.species_id = s.id) THEN 10 ELSE 0 END +
        CASE WHEN EXISTS (SELECT 1 FROM evidences e WHERE e.species_id = s.id) THEN 15 ELSE 0 END
    )::NUMERIC(5,2)
    FROM species s
    WHERE s.id = target_species_id;
$$ LANGUAGE SQL STABLE;

CREATE OR REPLACE FUNCTION refresh_species_data_quality()
RETURNS TRIGGER AS $$
DECLARE
    target_id UUID;
BEGIN
    target_id := CASE WHEN TG_OP = 'DELETE' THEN OLD.species_id ELSE NEW.species_id END;
    UPDATE species
    SET data_quality_score = COALESCE(calculate_species_data_quality(target_id), 0)
    WHERE id = target_id;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION refresh_own_species_data_quality()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE species
    SET data_quality_score = COALESCE(calculate_species_data_quality(NEW.id), 0)
    WHERE id = NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_species_quality_after_write
AFTER INSERT OR UPDATE OF latin_name, chinese_name, strain_number, source_environment, safety_level, summary
ON species FOR EACH ROW EXECUTE FUNCTION refresh_own_species_data_quality();

CREATE TRIGGER trg_aliases_species_quality
AFTER INSERT OR UPDATE OR DELETE ON species_aliases
FOR EACH ROW EXECUTE FUNCTION refresh_species_data_quality();

CREATE TRIGGER trg_functions_species_quality
AFTER INSERT OR UPDATE OR DELETE ON species_functions
FOR EACH ROW EXECUTE FUNCTION refresh_species_data_quality();

CREATE TRIGGER trg_culture_species_quality
AFTER INSERT OR UPDATE OR DELETE ON culture_conditions
FOR EACH ROW EXECUTE FUNCTION refresh_species_data_quality();

CREATE TRIGGER trg_evidences_species_quality
AFTER INSERT OR UPDATE OR DELETE ON evidences
FOR EACH ROW EXECUTE FUNCTION refresh_species_data_quality();

UPDATE species
SET data_quality_score = calculate_species_data_quality(id);
