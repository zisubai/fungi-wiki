CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE taxonomies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kingdom VARCHAR(120),
    phylum VARCHAR(120),
    class_name VARCHAR(120),
    order_name VARCHAR(120),
    family VARCHAR(120),
    genus VARCHAR(120),
    species_epithet VARCHAR(120),
    full_path TEXT,
    source VARCHAR(120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE species (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(160) NOT NULL UNIQUE,
    latin_name VARCHAR(255) NOT NULL,
    chinese_name VARCHAR(255),
    taxonomy_id UUID REFERENCES taxonomies(id),
    strain_number VARCHAR(160),
    source_environment VARCHAR(255),
    safety_level VARCHAR(80),
    is_model_organism BOOLEAN NOT NULL DEFAULT FALSE,
    summary TEXT,
    status VARCHAR(40) NOT NULL DEFAULT 'draft',
    data_quality_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    created_by UUID,
    updated_by UUID,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT species_status_check CHECK (status IN ('draft', 'pending_review', 'published', 'archived'))
);

CREATE TABLE species_aliases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    alias_name VARCHAR(255) NOT NULL,
    alias_type VARCHAR(80) NOT NULL DEFAULT 'synonym',
    source VARCHAR(120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE function_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES function_tags(id),
    name VARCHAR(120) NOT NULL,
    code VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE species_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    function_tag_id UUID NOT NULL REFERENCES function_tags(id),
    description TEXT,
    function_strength VARCHAR(80),
    verification_method VARCHAR(160),
    applicable_environment VARCHAR(255),
    confidence_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (species_id, function_tag_id)
);

CREATE TABLE culture_conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    medium_name VARCHAR(255),
    temperature_min NUMERIC(5,2),
    temperature_max NUMERIC(5,2),
    ph_min NUMERIC(4,2),
    ph_max NUMERIC(4,2),
    salinity_min NUMERIC(6,2),
    salinity_max NUMERIC(6,2),
    oxygen_requirement VARCHAR(120),
    culture_time VARCHAR(120),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE literatures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    authors TEXT,
    journal VARCHAR(255),
    publication_year INTEGER,
    doi VARCHAR(255),
    pmid VARCHAR(80),
    patent_number VARCHAR(120),
    source_url TEXT,
    abstract TEXT,
    source_type VARCHAR(80) NOT NULL DEFAULT 'paper',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT literatures_source_type_check CHECK (source_type IN ('paper', 'patent', 'database', 'case', 'internal'))
);

CREATE TABLE evidences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    function_tag_id UUID REFERENCES function_tags(id),
    literature_id UUID REFERENCES literatures(id),
    conclusion TEXT NOT NULL,
    evidence_level VARCHAR(40) NOT NULL DEFAULT 'medium',
    evidence_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT evidences_level_check CHECK (evidence_level IN ('low', 'medium', 'high', 'expert_verified'))
);

CREATE TABLE application_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    industry VARCHAR(120) NOT NULL,
    scenario VARCHAR(255) NOT NULL,
    problem TEXT,
    solution TEXT,
    result_summary TEXT,
    maturity_level VARCHAR(80),
    source VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(80) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(80) NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'pending',
    submitter_id UUID,
    reviewer_id UUID,
    comment TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    CONSTRAINT audit_records_status_check CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE TABLE search_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    query TEXT NOT NULL,
    filters JSONB NOT NULL DEFAULT '{}'::JSONB,
    result_count INTEGER NOT NULL DEFAULT 0,
    clicked_species_id UUID REFERENCES species(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE recommendation_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    requirement TEXT NOT NULL,
    parsed_intent JSONB NOT NULL DEFAULT '{}'::JSONB,
    recommended_species JSONB NOT NULL DEFAULT '[]'::JSONB,
    evidence_refs JSONB NOT NULL DEFAULT '[]'::JSONB,
    model_name VARCHAR(120),
    risk_level VARCHAR(80),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    entity_type VARCHAR(80) NOT NULL,
    entity_id UUID,
    feedback_type VARCHAR(80) NOT NULL,
    content TEXT,
    status VARCHAR(40) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_feedback_status_check CHECK (status IN ('open', 'processing', 'resolved', 'ignored'))
);

CREATE INDEX idx_species_latin_name ON species (latin_name);
CREATE INDEX idx_species_chinese_name ON species (chinese_name);
CREATE INDEX idx_species_status ON species (status);
CREATE INDEX idx_species_aliases_species_id ON species_aliases (species_id);
CREATE INDEX idx_species_aliases_alias_name ON species_aliases (alias_name);
CREATE INDEX idx_function_tags_parent_id ON function_tags (parent_id);
CREATE INDEX idx_species_functions_species_id ON species_functions (species_id);
CREATE INDEX idx_species_functions_function_tag_id ON species_functions (function_tag_id);
CREATE INDEX idx_culture_conditions_species_id ON culture_conditions (species_id);
CREATE INDEX idx_literatures_doi ON literatures (doi);
CREATE INDEX idx_literatures_pmid ON literatures (pmid);
CREATE INDEX idx_evidences_species_id ON evidences (species_id);
CREATE INDEX idx_evidences_function_tag_id ON evidences (function_tag_id);
CREATE INDEX idx_application_cases_species_id ON application_cases (species_id);
CREATE INDEX idx_audit_records_entity ON audit_records (entity_type, entity_id);
CREATE INDEX idx_search_logs_created_at ON search_logs (created_at);
CREATE INDEX idx_recommendation_records_created_at ON recommendation_records (created_at);
CREATE INDEX idx_user_feedback_entity ON user_feedback (entity_type, entity_id);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_taxonomies_updated_at BEFORE UPDATE ON taxonomies FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_species_updated_at BEFORE UPDATE ON species FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_function_tags_updated_at BEFORE UPDATE ON function_tags FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_species_functions_updated_at BEFORE UPDATE ON species_functions FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_culture_conditions_updated_at BEFORE UPDATE ON culture_conditions FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_literatures_updated_at BEFORE UPDATE ON literatures FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_evidences_updated_at BEFORE UPDATE ON evidences FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_application_cases_updated_at BEFORE UPDATE ON application_cases FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_user_feedback_updated_at BEFORE UPDATE ON user_feedback FOR EACH ROW EXECUTE FUNCTION set_updated_at();


INSERT INTO function_tags (name, code, description, sort_order)
VALUES
    ('促生', 'plant-growth-promotion', '促进植物生长，包括促根、促苗、提高养分吸收等。', 10),
    ('生防', 'biocontrol', '抑制病原微生物或降低病害发生。', 20),
    ('固氮', 'nitrogen-fixation', '将大气氮转化为可利用氮素。', 30),
    ('解磷', 'phosphate-solubilization', '溶解难溶性磷，提高磷素有效性。', 40),
    ('降解', 'biodegradation', '降解污染物、农残、石油烃或其他有机物。', 50),
    ('发酵', 'fermentation', '用于食品、工业或生物制造发酵过程。', 60),
    ('产酶', 'enzyme-production', '产生蛋白酶、淀粉酶、纤维素酶等工业酶。', 70)
ON CONFLICT (code) DO NOTHING;

INSERT INTO species (slug, latin_name, chinese_name, safety_level, is_model_organism, summary, status, published_at)
VALUES
    ('bacillus-subtilis', 'Bacillus subtilis', '枯草芽孢杆菌', 'BSL-1', TRUE, '常见功能菌和工业底盘菌，适合首批百科与功能菌数据库建设。', 'published', NOW()),
    ('saccharomyces-cerevisiae', 'Saccharomyces cerevisiae', '酿酒酵母', 'BSL-1', TRUE, '经典真核模式微生物，广泛用于发酵生产和合成生物学设计。', 'published', NOW())
ON CONFLICT (slug) DO NOTHING;
