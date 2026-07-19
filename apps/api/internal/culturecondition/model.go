package culturecondition

import "time"

type Condition struct {
	ID                string    `json:"id"`
	SpeciesID         string    `json:"speciesId"`
	MediumName        string    `json:"mediumName"`
	TemperatureMin    *float64  `json:"temperatureMin"`
	TemperatureMax    *float64  `json:"temperatureMax"`
	PHMin             *float64  `json:"phMin"`
	PHMax             *float64  `json:"phMax"`
	SalinityMin       *float64  `json:"salinityMin"`
	SalinityMax       *float64  `json:"salinityMax"`
	OxygenRequirement string    `json:"oxygenRequirement"`
	CultureTime       string    `json:"cultureTime"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Input struct {
	MediumName        string   `json:"mediumName"`
	TemperatureMin    *float64 `json:"temperatureMin"`
	TemperatureMax    *float64 `json:"temperatureMax"`
	PHMin             *float64 `json:"phMin"`
	PHMax             *float64 `json:"phMax"`
	SalinityMin       *float64 `json:"salinityMin"`
	SalinityMax       *float64 `json:"salinityMax"`
	OxygenRequirement string   `json:"oxygenRequirement"`
	CultureTime       string   `json:"cultureTime"`
	Notes             string   `json:"notes"`
}

type ReplaceInput struct {
	Items []Input `json:"items" binding:"required"`
}
