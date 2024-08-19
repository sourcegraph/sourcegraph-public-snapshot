package entities

// DiffusionCommit represents a commit in Diffusion.
type DiffusionCommit struct {
	ID             string `json:"id"`
	PHID           string `json:"phid"`
	RepositoryPHID string `json:"repositoryPHID"`
	Identifier     string `json:"identifier"`
	Epoch          string `json:"epoch"`
	URI            string `json:"uri"`
	IsImporting    bool   `json:"isImporting"`
	Summary        string `json:"summary"`
	AuthorPHID     string `json:"authorPHID"`
	CommitterPHID  string `json:"committerPHID"`
	Author         string `json:"author"`
	AuthorName     string `json:"authorName"`
	AuthorEmail    string `json:"authorEmail"`
	Committer      string `json:"committer"`
	CommitterName  string `json:"committerName"`
}
