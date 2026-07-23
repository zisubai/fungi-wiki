package smartsearch

import (
	"fungi-wiki/apps/api/internal/species"
	"time"
)

type ResultItem struct {
	species.Species
	HybridScore  float64  `json:"hybridScore"`
	MatchReasons []string `json:"matchReasons"`
}
type Result struct {
	Items           []ResultItem `json:"items"`
	Total           int          `json:"total"`
	Limit           int          `json:"limit"`
	Offset          int          `json:"offset"`
	SemanticEnabled bool         `json:"semanticEnabled"`
	ExpandedTerms   []string     `json:"expandedTerms"`
}
type Params struct {
	Query, FunctionTag, SafetyLevel, SourceEnvironment, Sort string
	Temperature, PH                                          *float64
	Limit, Offset                                            int
}
type Synonym struct {
	ID        string    `json:"id"`
	Term      string    `json:"term"`
	Value     string    `json:"synonym"`
	Weight    float64   `json:"weight"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Rule struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	QueryPattern    string    `json:"queryPattern"`
	FunctionTagCode string    `json:"functionTagCode"`
	SafetyLevel     string    `json:"safetyLevel"`
	Boost           float64   `json:"boost"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
type SynonymInput struct {
	Term    string  `json:"term" binding:"required"`
	Value   string  `json:"synonym" binding:"required"`
	Weight  float64 `json:"weight"`
	Enabled *bool   `json:"enabled"`
}
type RuleInput struct {
	Name            string  `json:"name" binding:"required"`
	QueryPattern    string  `json:"queryPattern" binding:"required"`
	FunctionTagCode string  `json:"functionTagCode"`
	SafetyLevel     string  `json:"safetyLevel"`
	Boost           float64 `json:"boost"`
	Enabled         *bool   `json:"enabled"`
}
