package searchanalytics

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Report(context.Context, int) (Report, error)
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }
func (r *PostgresRepository) Report(ctx context.Context, days int) (Report, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	report := Report{Days: days, PopularQueries: []QueryStat{}, NoResultQueries: []QueryStat{}}
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*),COUNT(*)FILTER(WHERE result_count=0),COUNT(DISTINCT NULLIF(query,'')) FROM search_logs WHERE created_at>=NOW()-($1*INTERVAL '1 day')`, days).Scan(&report.TotalSearches, &report.NoResultSearches, &report.DistinctQueries)
	if err != nil {
		return report, err
	}
	popular, err := r.pool.Query(ctx, `SELECT query,COUNT(*),AVG(result_count) FROM search_logs WHERE created_at>=NOW()-($1*INTERVAL '1 day') AND query<>'' GROUP BY query ORDER BY COUNT(*)DESC,query LIMIT 10`, days)
	if err != nil {
		return report, err
	}
	defer popular.Close()
	for popular.Next() {
		var x QueryStat
		if err = popular.Scan(&x.Query, &x.Count, &x.AverageResults); err != nil {
			return report, err
		}
		report.PopularQueries = append(report.PopularQueries, x)
	}
	empty, err := r.pool.Query(ctx, `SELECT query,COUNT(*),AVG(result_count) FROM search_logs WHERE created_at>=NOW()-($1*INTERVAL '1 day') AND result_count=0 AND query<>'' GROUP BY query ORDER BY COUNT(*)DESC,query LIMIT 10`, days)
	if err != nil {
		return report, err
	}
	defer empty.Close()
	for empty.Next() {
		var x QueryStat
		if err = empty.Scan(&x.Query, &x.Count, &x.AverageResults); err != nil {
			return report, err
		}
		report.NoResultQueries = append(report.NoResultQueries, x)
	}
	return report, empty.Err()
}
