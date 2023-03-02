package types

type RepoPathRanks struct {
	Statistics ReferenceCountStatistics `json:"statistics"`
	Paths      map[string]PathRank      `json:"paths"`
}

type ReferenceCountStatistics struct {
	Min               int     `json:"min"`
	Mean              float64 `json:"mean"`
	MaxReferenceCount int     `json:"max"`
}

type PathRank struct {
	ReferenceCount int `json:"reference_count"`
}
