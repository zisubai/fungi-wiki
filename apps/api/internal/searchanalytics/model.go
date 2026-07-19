package searchanalytics

type QueryStat struct {
	Query          string  `json:"query"`
	Count          int     `json:"count"`
	AverageResults float64 `json:"averageResults"`
}
type Report struct {
	Days             int         `json:"days"`
	TotalSearches    int         `json:"totalSearches"`
	NoResultSearches int         `json:"noResultSearches"`
	DistinctQueries  int         `json:"distinctQueries"`
	PopularQueries   []QueryStat `json:"popularQueries"`
	NoResultQueries  []QueryStat `json:"noResultQueries"`
}
