package types

type RepoPathRanks struct {
	MeanRank float64            `json:"mean_reference_count"`
	Paths    map[string]float64 `json:"paths"`
}
