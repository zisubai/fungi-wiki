package audit

import "time"

type Record struct {
	ID          string     `json:"id"`
	EntityType  string     `json:"entityType"`
	EntityID    string     `json:"entityId"`
	EntityName  string     `json:"entityName"`
	Action      string     `json:"action"`
	Status      string     `json:"status"`
	Comment     string     `json:"comment"`
	SubmittedAt time.Time  `json:"submittedAt"`
	ReviewedAt  *time.Time `json:"reviewedAt,omitempty"`
}
type ReviewInput struct {
	Comment string `json:"comment"`
}
