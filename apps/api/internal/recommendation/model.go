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
	ID                 string              `json:"id"`
	Slug               string              `json:"slug"`
	LatinName          string              `json:"latinName"`
	ChineseName        string              `json:"chineseName"`
	SafetyLevel        string              `json:"safetyLevel"`
	Summary            string              `json:"summary"`
	Score              float64             `json:"score"`
	EvidenceCount      int                 `json:"evidenceCount"`
	EvidenceReferences []EvidenceReference `json:"evidenceReferences"`
	Reasons            []string            `json:"reasons"`
	RiskWarning        string              `json:"riskWarning,omitempty"`
}
type EvidenceReference struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	PublicationYear *int    `json:"publicationYear,omitempty"`
	DOI             string  `json:"doi,omitempty"`
	PMID            string  `json:"pmid,omitempty"`
	SourceURL       string  `json:"sourceUrl,omitempty"`
	Conclusion      string  `json:"conclusion"`
	EvidenceLevel   string  `json:"evidenceLevel"`
	EvidenceScore   float64 `json:"evidenceScore"`
}
type Response struct {
	RecordID              string         `json:"recordId"`
	ParsedFunctionTag     string         `json:"parsedFunctionTag,omitempty"`
	ParsedIntent          map[string]any `json:"parsedIntent"`
	Items                 []Item         `json:"items"`
	RelaxationSuggestions []string       `json:"relaxationSuggestions,omitempty"`
	Disclaimer            string         `json:"disclaimer"`
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
	Total        int                        `json:"total"`
	Helpful      int                        `json:"helpful"`
	Unhelpful    int                        `json:"unhelpful"`
	Records      []QualityRecord            `json:"records"`
	Combinations []CombinationQualityRecord `json:"combinations"`
}

type CombinationQualityRecord struct {
	ID             string                  `json:"id"`
	FunctionTags   []string                `json:"functionTags"`
	SafetyLevel    string                  `json:"safetyLevel"`
	Items          []Combination           `json:"items"`
	ModelName      string                  `json:"modelName"`
	RiskLevel      string                  `json:"riskLevel"`
	HelpfulCount   int                     `json:"helpfulCount"`
	UnhelpfulCount int                     `json:"unhelpfulCount"`
	Experiments    []CombinationExperiment `json:"experiments"`
	CreatedAt      time.Time               `json:"createdAt"`
}

type CombinationExperimentInput struct {
	CandidateIndex *int     `json:"candidateIndex" binding:"required"`
	Outcome        string   `json:"outcome" binding:"required,oneof=compatible incompatible inconclusive"`
	Temperature    *float64 `json:"temperature"`
	PH             *float64 `json:"ph"`
	Notes          string   `json:"notes"`
}

type CombinationExperiment struct {
	ID                  string              `json:"id"`
	CombinationRecordID string              `json:"combinationRecordId"`
	CandidateIndex      int                 `json:"candidateIndex"`
	CandidateMembers    []CombinationMember `json:"candidateMembers"`
	Outcome             string              `json:"outcome"`
	Temperature         *float64            `json:"temperature,omitempty"`
	PH                  *float64            `json:"ph,omitempty"`
	Notes               string              `json:"notes"`
	CreatedAt           time.Time           `json:"createdAt"`
}

type CombinationInput struct {
	FunctionTags []string `json:"functionTags" binding:"required,len=2,dive,required"`
	SafetyLevel  string   `json:"safetyLevel"`
}

type CombinationMember struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	LatinName     string   `json:"latinName"`
	ChineseName   string   `json:"chineseName"`
	SafetyLevel   string   `json:"safetyLevel"`
	FunctionTags  []string `json:"functionTags"`
	EvidenceCount int      `json:"evidenceCount"`
}

type Combination struct {
	Members        []CombinationMember `json:"members"`
	Score          float64             `json:"score"`
	TemperatureMin *float64            `json:"temperatureMin"`
	TemperatureMax *float64            `json:"temperatureMax"`
	PHMin          *float64            `json:"phMin"`
	PHMax          *float64            `json:"phMax"`
	Compatible     bool                `json:"compatible"`
	Reasons        []string            `json:"reasons"`
	Warning        string              `json:"warning,omitempty"`
}

type CombinationResponse struct {
	RecordID   string        `json:"recordId"`
	Items      []Combination `json:"items"`
	Disclaimer string        `json:"disclaimer"`
}
