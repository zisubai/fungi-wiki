package speciesalias

import "time"

type Alias struct {
	ID        string    `json:"id"`
	SpeciesID string    `json:"speciesId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"createdAt"`
}
type Input struct {
	Name   string `json:"name" binding:"required"`
	Type   string `json:"type"`
	Source string `json:"source"`
}
type ReplaceInput struct {
	Items []Input `json:"items" binding:"required"`
}
