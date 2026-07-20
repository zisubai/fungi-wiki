package dataquality

type MissingStat struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type PrioritySpecies struct {
	ID        string  `json:"id"`
	Slug      string  `json:"slug"`
	LatinName string  `json:"latinName"`
	Status    string  `json:"status"`
	Score     float64 `json:"score"`
}

type Report struct {
	Total           int               `json:"total"`
	AverageScore    float64           `json:"averageScore"`
	Complete        int               `json:"complete"`
	NeedsCompletion int               `json:"needsCompletion"`
	Incomplete      int               `json:"incomplete"`
	Missing         []MissingStat     `json:"missing"`
	PrioritySpecies []PrioritySpecies `json:"prioritySpecies"`
}
