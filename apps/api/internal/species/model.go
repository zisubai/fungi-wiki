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
	Query  string
	Status string
	Limit  int
	Offset int
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
