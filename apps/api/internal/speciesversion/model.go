package speciesversion

import (
	"encoding/json"
	"time"
)

type Version struct {
	ID            string          `json:"id"`
	SpeciesID     string          `json:"speciesId"`
	VersionNumber int             `json:"versionNumber"`
	ChangeType    string          `json:"changeType"`
	SourceTable   string          `json:"sourceTable"`
	Snapshot      json.RawMessage `json:"snapshot"`
	CreatedAt     time.Time       `json:"createdAt"`
}
