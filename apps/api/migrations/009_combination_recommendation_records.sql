CREATE TABLE combination_recommendation_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    safety_level VARCHAR(80),
    combinations JSONB NOT NULL DEFAULT '[]'::JSONB,
    model_name VARCHAR(120) NOT NULL DEFAULT 'combination-rules-v2',
    risk_level VARCHAR(80) NOT NULL DEFAULT 'review_required',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_combination_recommendation_records_created_at
ON combination_recommendation_records (created_at DESC);
