CREATE TABLE species_embeddings (
    species_id UUID PRIMARY KEY REFERENCES species(id) ON DELETE CASCADE,
    model VARCHAR(160) NOT NULL,
    dimensions INTEGER NOT NULL,
    embedding REAL[] NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    embedded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('member','operator','expert','admin'));

CREATE TABLE search_synonyms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    term VARCHAR(160) NOT NULL,
    synonym VARCHAR(160) NOT NULL,
    weight NUMERIC(4,2) NOT NULL DEFAULT 0.85,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(term, synonym),
    CONSTRAINT search_synonym_weight CHECK(weight > 0 AND weight <= 2)
);

CREATE TABLE search_recall_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(160) NOT NULL,
    query_pattern VARCHAR(255) NOT NULL,
    function_tag_code VARCHAR(120),
    safety_level VARCHAR(80),
    boost NUMERIC(5,2) NOT NULL DEFAULT 0.15,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT search_rule_boost CHECK(boost >= 0 AND boost <= 5)
);

CREATE TABLE user_species_favorites (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(user_id, species_id)
);

CREATE TABLE user_search_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    filters JSONB NOT NULL DEFAULT '{}'::JSONB,
    result_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_search_synonyms_term ON search_synonyms(lower(term)) WHERE enabled;
CREATE INDEX idx_search_rules_enabled ON search_recall_rules(enabled);
CREATE INDEX idx_user_search_history_user_created ON user_search_history(user_id, created_at DESC);

CREATE TRIGGER trg_search_synonyms_updated_at BEFORE UPDATE ON search_synonyms FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_search_recall_rules_updated_at BEFORE UPDATE ON search_recall_rules FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION cosine_similarity(left_vector REAL[], right_vector REAL[])
RETURNS DOUBLE PRECISION AS $$
DECLARE
    dot_product DOUBLE PRECISION := 0;
    left_norm DOUBLE PRECISION := 0;
    right_norm DOUBLE PRECISION := 0;
    vector_index INTEGER;
BEGIN
    IF left_vector IS NULL OR right_vector IS NULL OR array_length(left_vector, 1) IS DISTINCT FROM array_length(right_vector, 1) THEN RETURN 0; END IF;
    FOR vector_index IN 1..array_length(left_vector, 1) LOOP
        dot_product := dot_product + left_vector[vector_index] * right_vector[vector_index];
        left_norm := left_norm + left_vector[vector_index] * left_vector[vector_index];
        right_norm := right_norm + right_vector[vector_index] * right_vector[vector_index];
    END LOOP;
    IF left_norm = 0 OR right_norm = 0 THEN RETURN 0; END IF;
    RETURN dot_product / (sqrt(left_norm) * sqrt(right_norm));
END;
$$ LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE;

INSERT INTO search_synonyms(term, synonym) VALUES
('促生', '植物生长促进'), ('生防', '病害防控'), ('降解', '污染物去除'), ('发酵', '生物制造')
ON CONFLICT DO NOTHING;
