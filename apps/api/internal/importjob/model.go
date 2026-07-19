package importjob

import "time"

type SpeciesRow struct {
	RowNumber         int      `json:"rowNumber"`
	Slug              string   `json:"slug"`
	LatinName         string   `json:"latinName"`
	ChineseName       string   `json:"chineseName"`
	StrainNumber      string   `json:"strainNumber"`
	SourceEnvironment string   `json:"sourceEnvironment"`
	SafetyLevel       string   `json:"safetyLevel"`
	IsModelOrganism   bool     `json:"isModelOrganism"`
	Summary           string   `json:"summary"`
	FunctionTags      []string `json:"functionTags"`
	MediumName        string   `json:"mediumName"`
	TemperatureMin    *float64 `json:"temperatureMin"`
	TemperatureMax    *float64 `json:"temperatureMax"`
	PHMin             *float64 `json:"phMin"`
	PHMax             *float64 `json:"phMax"`
	OxygenRequirement string   `json:"oxygenRequirement"`
	CultureTime       string   `json:"cultureTime"`
	Errors            []string `json:"errors,omitempty"`
}

type RowResult struct {
	RowNumber int      `json:"rowNumber"`
	Slug      string   `json:"slug"`
	Status    string   `json:"status"`
	Errors    []string `json:"errors,omitempty"`
}

type Batch struct {
	ID             string      `json:"id"`
	SourceFilename string      `json:"sourceFilename"`
	TotalRows      int         `json:"totalRows"`
	SuccessRows    int         `json:"successRows"`
	FailedRows     int         `json:"failedRows"`
	Status         string      `json:"status"`
	CreatedAt      time.Time   `json:"createdAt"`
	CompletedAt    *time.Time  `json:"completedAt,omitempty"`
	Rows           []RowResult `json:"rows,omitempty"`
}
