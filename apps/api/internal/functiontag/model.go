package functiontag

import "time"

type FunctionTag struct {
	ID                    string    `json:"id"`
	ParentID              string    `json:"parentId"`
	Name                  string    `json:"name"`
	Code                  string    `json:"code"`
	Description           string    `json:"description"`
	SortOrder             int       `json:"sortOrder"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
	PublishedSpeciesCount int       `json:"publishedSpeciesCount"`
}

type ListParams struct {
	Query  string
	Limit  int
	Offset int
}

type CreateInput struct {
	ParentID    string `json:"parentId"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

type UpdateInput struct {
	ParentID    string `json:"parentId"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}
