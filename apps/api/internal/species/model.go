package species

import "time"

type Status string

const (
	StatusDraft         Status = "draft"
	StatusPendingReview Status = "pending_review"
	StatusPublished     Status = "published"
	StatusArchived      Status = "archived"
)

type Species struct {
	ID                string     `json:"id"`
	Slug              string     `json:"slug"`
	LatinName         string     `json:"latinName"`
	ChineseName       string     `json:"chineseName"`
	StrainNumber      string     `json:"strainNumber"`
	SourceEnvironment string     `json:"sourceEnvironment"`
	SafetyLevel       string     `json:"safetyLevel"`
	IsModelOrganism   bool       `json:"isModelOrganism"`
	Summary           string     `json:"summary"`
	Status            Status     `json:"status"`
	DataQualityScore  float64    `json:"dataQualityScore"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	PublishedAt       *time.Time `json:"publishedAt,omitempty"`
}

type ListParams struct {
	Query             string
	Status            string
	FunctionTag       string
	Temperature       *float64
	PH                *float64
	SafetyLevel       string
	SourceEnvironment string
	Sort              string
	Limit             int
	Offset            int
}

type ListResult struct {
	Items []Species
	Total int
}

type Comparison struct {
	Species
	FunctionTags   []string `json:"functionTags"`
	TemperatureMin *float64 `json:"temperatureMin"`
	TemperatureMax *float64 `json:"temperatureMax"`
	PHMin          *float64 `json:"phMin"`
	PHMax          *float64 `json:"phMax"`
	EvidenceCount  int      `json:"evidenceCount"`
}

type QualityComponent struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Weight    int    `json:"weight"`
	Completed bool   `json:"completed"`
}

type QualityReport struct {
	Score      float64            `json:"score"`
	Components []QualityComponent `json:"components"`
}

type CreateInput struct {
	Slug              string `json:"slug" binding:"required"`
	LatinName         string `json:"latinName" binding:"required"`
	ChineseName       string `json:"chineseName"`
	StrainNumber      string `json:"strainNumber"`
	SourceEnvironment string `json:"sourceEnvironment"`
	SafetyLevel       string `json:"safetyLevel"`
	IsModelOrganism   bool   `json:"isModelOrganism"`
	Summary           string `json:"summary"`
	Status            Status `json:"status"`
}

type UpdateInput struct {
	Slug              string `json:"slug" binding:"required"`
	LatinName         string `json:"latinName" binding:"required"`
	ChineseName       string `json:"chineseName"`
	StrainNumber      string `json:"strainNumber"`
	SourceEnvironment string `json:"sourceEnvironment"`
	SafetyLevel       string `json:"safetyLevel"`
	IsModelOrganism   bool   `json:"isModelOrganism"`
	Summary           string `json:"summary"`
	Status            Status `json:"status"`
}
