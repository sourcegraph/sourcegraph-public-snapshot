package types

// RepoPathRanks are given to Zoekt when a repository has precise reference counts.
type RepoPathRanks struct {
	// MeanRank is the binary log mean of references counts over all repositories.
	MeanRank float64 `json:"mean_reference_count"`

	// Paths are a map from path name to the number of references for a symbol defined
	// in that path for a particular repository. Counts include cross-repository refs.
	Paths map[string]float64 `json:"paths"`
}
