package compute

type Text struct {
	Value        string `json:"value"`
	Kind         string `json:"kind"`
	RepositoryID int32  `json:"repositoryID"`
	Repository   string `json:"repository"`
}
