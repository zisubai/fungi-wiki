package evidence

import "time"

type Evidence struct {
	ID              string    `json:"id"`
	SpeciesID       string    `json:"speciesId"`
	FunctionTagID   string    `json:"functionTagId"`
	FunctionTagName string    `json:"functionTagName"`
	LiteratureID    string    `json:"literatureId"`
	Title           string    `json:"title"`
	Authors         string    `json:"authors"`
	Journal         string    `json:"journal"`
	PublicationYear *int      `json:"publicationYear"`
	DOI             string    `json:"doi"`
	PMID            string    `json:"pmid"`
	SourceURL       string    `json:"sourceUrl"`
	Abstract        string    `json:"abstract"`
	Conclusion      string    `json:"conclusion"`
	EvidenceLevel   string    `json:"evidenceLevel"`
	EvidenceScore   float64   `json:"evidenceScore"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
type CreateInput struct {
	FunctionTagID   string  `json:"functionTagId"`
	Title           string  `json:"title" binding:"required"`
	Authors         string  `json:"authors"`
	Journal         string  `json:"journal"`
	PublicationYear *int    `json:"publicationYear"`
	DOI             string  `json:"doi"`
	PMID            string  `json:"pmid"`
	SourceURL       string  `json:"sourceUrl"`
	Abstract        string  `json:"abstract"`
	Conclusion      string  `json:"conclusion" binding:"required"`
	EvidenceLevel   string  `json:"evidenceLevel"`
	EvidenceScore   float64 `json:"evidenceScore"`
}
