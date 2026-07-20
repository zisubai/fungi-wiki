package dataquality

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Report(context.Context) (Report, error)
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repo *PostgresRepository) Report(ctx context.Context) (Report, error) {
	report := Report{Missing: []MissingStat{}, PrioritySpecies: []PrioritySpecies{}}
	err := repo.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(AVG(data_quality_score), 0),
		       COUNT(*) FILTER (WHERE data_quality_score >= 80),
		       COUNT(*) FILTER (WHERE data_quality_score >= 60 AND data_quality_score < 80),
		       COUNT(*) FILTER (WHERE data_quality_score < 60)
		FROM species WHERE status <> 'archived'
	`).Scan(&report.Total, &report.AverageScore, &report.Complete, &report.NeedsCompletion, &report.Incomplete)
	if err != nil {
		return report, err
	}

	var strain, source, safety, summary, aliases, functions, culture, evidences int
	err = repo.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE NULLIF(BTRIM(strain_number), '') IS NULL),
		       COUNT(*) FILTER (WHERE NULLIF(BTRIM(source_environment), '') IS NULL),
		       COUNT(*) FILTER (WHERE NULLIF(BTRIM(safety_level), '') IS NULL),
		       COUNT(*) FILTER (WHERE NULLIF(BTRIM(summary), '') IS NULL),
		       COUNT(*) FILTER (WHERE NOT EXISTS(SELECT 1 FROM species_aliases sa WHERE sa.species_id=species.id)),
		       COUNT(*) FILTER (WHERE NOT EXISTS(SELECT 1 FROM species_functions sf WHERE sf.species_id=species.id)),
		       COUNT(*) FILTER (WHERE NOT EXISTS(SELECT 1 FROM culture_conditions cc WHERE cc.species_id=species.id)),
		       COUNT(*) FILTER (WHERE NOT EXISTS(SELECT 1 FROM evidences e WHERE e.species_id=species.id))
		FROM species WHERE status <> 'archived'
	`).Scan(&strain, &source, &safety, &summary, &aliases, &functions, &culture, &evidences)
	if err != nil {
		return report, err
	}
	report.Missing = []MissingStat{
		{Key: "strainNumber", Label: "保藏/菌株编号", Count: strain}, {Key: "sourceEnvironment", Label: "来源环境", Count: source},
		{Key: "safetyLevel", Label: "安全等级", Count: safety}, {Key: "summary", Label: "菌种摘要", Count: summary},
		{Key: "aliases", Label: "别名与同义词", Count: aliases}, {Key: "functions", Label: "功能标签", Count: functions},
		{Key: "cultureConditions", Label: "培养条件", Count: culture}, {Key: "evidences", Label: "文献证据", Count: evidences},
	}
	sort.SliceStable(report.Missing, func(i, j int) bool { return report.Missing[i].Count > report.Missing[j].Count })

	rows, err := repo.pool.Query(ctx, `SELECT id::text,slug,latin_name,status,data_quality_score FROM species WHERE status <> 'archived' ORDER BY data_quality_score,updated_at LIMIT 20`)
	if err != nil {
		return report, err
	}
	defer rows.Close()
	for rows.Next() {
		var item PrioritySpecies
		if err = rows.Scan(&item.ID, &item.Slug, &item.LatinName, &item.Status, &item.Score); err != nil {
			return report, err
		}
		report.PrioritySpecies = append(report.PrioritySpecies, item)
	}
	return report, rows.Err()
}
