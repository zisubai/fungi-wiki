package recommendation

import "time"

type Input struct {
	Requirement       string   `json:"requirement" binding:"required"`
	FunctionTag       string   `json:"functionTag"`
	Temperature       *float64 `json:"temperature"`
	PH                *float64 `json:"ph"`
	SafetyLevel       string   `json:"safetyLevel"`
	SourceEnvironment string   `json:"sourceEnvironment"`
	Limit             int      `json:"limit"`
}
type Item struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	LatinName     string   `json:"latinName"`
	ChineseName   string   `json:"chineseName"`
	SafetyLevel   string   `json:"safetyLevel"`
	Summary       string   `json:"summary"`
	Score         float64  `json:"score"`
	EvidenceCount int      `json:"evidenceCount"`
	Reasons       []string `json:"reasons"`
	RiskWarning   string   `json:"riskWarning,omitempty"`
}
type Response struct {
	RecordID          string `json:"recordId"`
	ParsedFunctionTag string `json:"parsedFunctionTag,omitempty"`
	Items             []Item `json:"items"`
	Disclaimer        string `json:"disclaimer"`
}
type FeedbackInput struct {
	FeedbackType string `json:"feedbackType" binding:"required,oneof=helpful unhelpful"`
	Content      string `json:"content"`
}
type QualityRecord struct {
	ID             string         `json:"id"`
	Requirement    string         `json:"requirement"`
	ParsedIntent   map[string]any `json:"parsedIntent"`
	Items          []Item         `json:"items"`
	ModelName      string         `json:"modelName"`
	RiskLevel      string         `json:"riskLevel"`
	HelpfulCount   int            `json:"helpfulCount"`
	UnhelpfulCount int            `json:"unhelpfulCount"`
	CreatedAt      time.Time      `json:"createdAt"`
}
type QualityReport struct {
	Total     int             `json:"total"`
	Helpful   int             `json:"helpful"`
	Unhelpful int             `json:"unhelpful"`
	Records   []QualityRecord `json:"records"`
}
