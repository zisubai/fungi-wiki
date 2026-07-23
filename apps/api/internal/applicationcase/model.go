package applicationcase

import "time"

type Case struct {
	ID            string    `json:"id"`
	SpeciesID     string    `json:"speciesId"`
	Industry      string    `json:"industry"`
	Scenario      string    `json:"scenario"`
	Problem       string    `json:"problem"`
	Solution      string    `json:"solution"`
	ResultSummary string    `json:"resultSummary"`
	MaturityLevel string    `json:"maturityLevel"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Input struct {
	Industry      string `json:"industry" binding:"required"`
	Scenario      string `json:"scenario" binding:"required"`
	Problem       string `json:"problem"`
	Solution      string `json:"solution"`
	ResultSummary string `json:"resultSummary"`
	MaturityLevel string `json:"maturityLevel"`
	Source        string `json:"source"`
}
