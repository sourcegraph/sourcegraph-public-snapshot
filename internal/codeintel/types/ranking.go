package types

type RepoPathRanks struct {
	Statistics         ReferenceCountStatistics `json:"statistics"`
	MeanReferenceCount float64                  `json:"mean_reference_count"`
	Paths              map[string]PathRank      `json:"paths"`
}

type ReferenceCountStatistics struct {
	MinReferenceCount  int     `json:"min_reference_count"`
	MeanReferenceCount float64 `json:"mean_reference_count"`
	MaxReferenceCount  int     `json:"max_reference_count"`
}

type PathRank struct {
	ReferenceCount int `json:"reference_count"`
}
