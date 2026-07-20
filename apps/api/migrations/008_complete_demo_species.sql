-- Complete the two built-in development records so local search, recommendation,
-- evidence, comparison, and quality workflows can be verified end to end.
UPDATE species
SET strain_number = COALESCE(strain_number, 'ATCC 6051'),
    source_environment = COALESCE(source_environment, '土壤')
WHERE slug = 'bacillus-subtilis';

UPDATE species
SET strain_number = COALESCE(strain_number, 'ATCC 204508'),
    source_environment = COALESCE(source_environment, '食品发酵')
WHERE slug = 'saccharomyces-cerevisiae';

INSERT INTO species_aliases(species_id, alias_name, alias_type, source)
SELECT id, 'B. subtilis', 'abbreviation', 'demo-seed'
FROM species WHERE slug = 'bacillus-subtilis'
ON CONFLICT DO NOTHING;

INSERT INTO species_aliases(species_id, alias_name, alias_type, source)
SELECT id, 'S. cerevisiae', 'abbreviation', 'demo-seed'
FROM species WHERE slug = 'saccharomyces-cerevisiae'
ON CONFLICT DO NOTHING;

INSERT INTO species_functions(species_id, function_tag_id, description, function_strength, verification_method, applicable_environment, confidence_score)
SELECT s.id, ft.id, '用于演示结构化功能菌检索和推荐。', 'high', 'literature', '土壤与发酵', 90
FROM species s JOIN function_tags ft ON ft.code = 'biocontrol'
WHERE s.slug = 'bacillus-subtilis'
ON CONFLICT (species_id, function_tag_id) DO NOTHING;

INSERT INTO species_functions(species_id, function_tag_id, description, function_strength, verification_method, applicable_environment, confidence_score)
SELECT s.id, ft.id, '用于演示发酵菌检索和推荐。', 'high', 'literature', '食品与工业发酵', 95
FROM species s JOIN function_tags ft ON ft.code = 'fermentation'
WHERE s.slug = 'saccharomyces-cerevisiae'
ON CONFLICT (species_id, function_tag_id) DO NOTHING;

INSERT INTO culture_conditions(species_id, medium_name, temperature_min, temperature_max, ph_min, ph_max, oxygen_requirement, culture_time, notes)
SELECT id, 'LB', 25, 37, 6, 8, '好氧', '18–24 h', '本地验收示例条件'
FROM species s
WHERE slug = 'bacillus-subtilis'
  AND NOT EXISTS (SELECT 1 FROM culture_conditions cc WHERE cc.species_id = s.id);

INSERT INTO culture_conditions(species_id, medium_name, temperature_min, temperature_max, ph_min, ph_max, oxygen_requirement, culture_time, notes)
SELECT id, 'YPD', 20, 30, 4, 6, '兼性厌氧', '24–48 h', '本地验收示例条件'
FROM species s
WHERE slug = 'saccharomyces-cerevisiae'
  AND NOT EXISTS (SELECT 1 FROM culture_conditions cc WHERE cc.species_id = s.id);

WITH inserted AS (
    INSERT INTO literatures(title, authors, journal, publication_year, doi, source_url, abstract, source_type)
    SELECT 'Bacillus subtilis as a model and industrial organism', 'Fungi Wiki Demo', 'Demo Evidence', 2026,
           '10.0000/fungi.bacillus.demo', 'https://example.com/fungi-wiki/bacillus', '用于本地功能演示的可替换文献记录。', 'paper'
    WHERE NOT EXISTS (SELECT 1 FROM literatures WHERE doi = '10.0000/fungi.bacillus.demo')
    RETURNING id
), literature AS (
    SELECT id FROM inserted
    UNION ALL
    SELECT id FROM literatures WHERE doi = '10.0000/fungi.bacillus.demo'
    LIMIT 1
)
INSERT INTO evidences(species_id, function_tag_id, literature_id, conclusion, evidence_level, evidence_score)
SELECT s.id, ft.id, literature.id, '示例证据支持枯草芽孢杆菌的生防应用候选价值。', 'medium', 70
FROM species s
JOIN function_tags ft ON ft.code = 'biocontrol'
CROSS JOIN literature
WHERE s.slug = 'bacillus-subtilis'
  AND NOT EXISTS (SELECT 1 FROM evidences e WHERE e.species_id=s.id AND e.literature_id=literature.id);

WITH inserted AS (
    INSERT INTO literatures(title, authors, journal, publication_year, doi, source_url, abstract, source_type)
    SELECT 'Saccharomyces cerevisiae fermentation reference', 'Fungi Wiki Demo', 'Demo Evidence', 2026,
           '10.0000/fungi.yeast.demo', 'https://example.com/fungi-wiki/yeast', '用于本地功能演示的可替换文献记录。', 'paper'
    WHERE NOT EXISTS (SELECT 1 FROM literatures WHERE doi = '10.0000/fungi.yeast.demo')
    RETURNING id
), literature AS (
    SELECT id FROM inserted
    UNION ALL
    SELECT id FROM literatures WHERE doi = '10.0000/fungi.yeast.demo'
    LIMIT 1
)
INSERT INTO evidences(species_id, function_tag_id, literature_id, conclusion, evidence_level, evidence_score)
SELECT s.id, ft.id, literature.id, '示例证据支持酿酒酵母的发酵应用价值。', 'medium', 75
FROM species s
JOIN function_tags ft ON ft.code = 'fermentation'
CROSS JOIN literature
WHERE s.slug = 'saccharomyces-cerevisiae'
  AND NOT EXISTS (SELECT 1 FROM evidences e WHERE e.species_id=s.id AND e.literature_id=literature.id);
