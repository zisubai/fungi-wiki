ALTER TABLE combination_experiments
    ADD COLUMN IF NOT EXISTS candidate_index INTEGER NOT NULL DEFAULT 0 CHECK (candidate_index >= 0),
    ADD COLUMN IF NOT EXISTS candidate_members JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE combination_experiments e
SET candidate_members = COALESCE(c.combinations -> e.candidate_index -> 'members', '[]'::jsonb)
FROM combination_recommendation_records c
WHERE c.id = e.combination_record_id
  AND e.candidate_members = '[]'::jsonb;
