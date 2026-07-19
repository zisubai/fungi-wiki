package speciesfunction

import "time"

type SpeciesFunction struct {
	ID                    string    `json:"id"`
	SpeciesID             string    `json:"speciesId"`
	FunctionTagID         string    `json:"functionTagId"`
	FunctionTagName       string    `json:"functionTagName"`
	FunctionTagCode       string    `json:"functionTagCode"`
	Description           string    `json:"description"`
	FunctionStrength      string    `json:"functionStrength"`
	VerificationMethod    string    `json:"verificationMethod"`
	ApplicableEnvironment string    `json:"applicableEnvironment"`
	ConfidenceScore       float64   `json:"confidenceScore"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type ReplaceItem struct {
	FunctionTagID         string  `json:"functionTagId" binding:"required"`
	Description           string  `json:"description"`
	FunctionStrength      string  `json:"functionStrength"`
	VerificationMethod    string  `json:"verificationMethod"`
	ApplicableEnvironment string  `json:"applicableEnvironment"`
	ConfidenceScore       float64 `json:"confidenceScore"`
}

type ReplaceInput struct {
	Items []ReplaceItem `json:"items" binding:"required"`
}
