package importjob

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Import(context.Context, string, []SpeciesRow) (Batch, error)
	List(context.Context, int) ([]Batch, error)
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool} }

func (repo *PostgresRepository) Import(ctx context.Context, filename string, rows []SpeciesRow) (Batch, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Batch{}, err
	}
	defer tx.Rollback(ctx)
	var batch Batch
	err = tx.QueryRow(ctx, `INSERT INTO import_batches(source_filename,total_rows) VALUES($1,$2) RETURNING id::text,source_filename,total_rows,success_rows,failed_rows,status,created_at,completed_at`, filename, len(rows)).Scan(&batch.ID, &batch.SourceFilename, &batch.TotalRows, &batch.SuccessRows, &batch.FailedRows, &batch.Status, &batch.CreatedAt, &batch.CompletedAt)
	if err != nil {
		return Batch{}, err
	}
	batch.Rows = make([]RowResult, 0, len(rows))
	for _, row := range rows {
		result := RowResult{RowNumber: row.RowNumber, Slug: row.Slug, Status: "failed", Errors: append([]string{}, row.Errors...)}
		var exists bool
		if len(result.Errors) == 0 {
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM species WHERE slug=$1)`, row.Slug).Scan(&exists); err != nil {
				return Batch{}, err
			}
			if exists {
				result.Errors = append(result.Errors, "slug 已存在")
			}
		}
		tagIDs := []string{}
		if len(result.Errors) == 0 {
			for _, code := range row.FunctionTags {
				var id string
				e := tx.QueryRow(ctx, `SELECT id::text FROM function_tags WHERE code=$1 OR name=$1`, code).Scan(&id)
				if e == pgx.ErrNoRows {
					result.Errors = append(result.Errors, fmt.Sprintf("功能标签不存在：%s", code))
					continue
				}
				if e != nil {
					return Batch{}, e
				}
				tagIDs = append(tagIDs, id)
			}
		}
		var speciesID string
		if len(result.Errors) == 0 {
			err = tx.QueryRow(ctx, `INSERT INTO species(slug,latin_name,chinese_name,strain_number,source_environment,safety_level,is_model_organism,summary,status,published_at) VALUES($1,$2,NULLIF($3,''),NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7,NULLIF($8,''),'pending_review',NULL) RETURNING id::text`, row.Slug, row.LatinName, row.ChineseName, row.StrainNumber, row.SourceEnvironment, row.SafetyLevel, row.IsModelOrganism, row.Summary).Scan(&speciesID)
			if err != nil {
				return Batch{}, err
			}
			for _, tagID := range tagIDs {
				if _, err = tx.Exec(ctx, `INSERT INTO species_functions(species_id,function_tag_id)VALUES($1::uuid,$2::uuid)`, speciesID, tagID); err != nil {
					return Batch{}, err
				}
			}
			if row.MediumName != "" || row.TemperatureMin != nil || row.TemperatureMax != nil || row.PHMin != nil || row.PHMax != nil || row.OxygenRequirement != "" || row.CultureTime != "" {
				if _, err = tx.Exec(ctx, `INSERT INTO culture_conditions(species_id,medium_name,temperature_min,temperature_max,ph_min,ph_max,oxygen_requirement,culture_time)VALUES($1::uuid,NULLIF($2,''),$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''))`, speciesID, row.MediumName, row.TemperatureMin, row.TemperatureMax, row.PHMin, row.PHMax, row.OxygenRequirement, row.CultureTime); err != nil {
					return Batch{}, err
				}
			}
			if _, err = tx.Exec(ctx, `INSERT INTO audit_records(entity_type,entity_id,action,status,comment)VALUES('species',$1::uuid,'import_publish','pending','批量导入，等待审核')`, speciesID); err != nil {
				return Batch{}, err
			}
			result.Status = "imported"
		}
		raw, _ := json.Marshal(row)
		errorMessage := strings.Join(result.Errors, "；")
		if _, err = tx.Exec(ctx, `INSERT INTO import_rows(batch_id,row_number,raw_data,species_id,status,error_message)VALUES($1::uuid,$2,$3::jsonb,NULLIF($4,'')::uuid,$5,NULLIF($6,''))`, batch.ID, row.RowNumber, string(raw), speciesID, result.Status, errorMessage); err != nil {
			return Batch{}, err
		}
		if result.Status == "imported" {
			batch.SuccessRows++
		} else {
			batch.FailedRows++
		}
		batch.Rows = append(batch.Rows, result)
	}
	_, err = tx.Exec(ctx, `UPDATE import_batches SET success_rows=$2,failed_rows=$3,status='completed',completed_at=NOW() WHERE id=$1::uuid`, batch.ID, batch.SuccessRows, batch.FailedRows)
	if err != nil {
		return Batch{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Batch{}, err
	}
	batch.Status = "completed"
	return batch, nil
}

func (repo *PostgresRepository) List(ctx context.Context, limit int) ([]Batch, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := repo.pool.Query(ctx, `SELECT id::text,source_filename,total_rows,success_rows,failed_rows,status,created_at,completed_at FROM import_batches ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Batch{}
	for rows.Next() {
		var x Batch
		if err = rows.Scan(&x.ID, &x.SourceFilename, &x.TotalRows, &x.SuccessRows, &x.FailedRows, &x.Status, &x.CreatedAt, &x.CompletedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
