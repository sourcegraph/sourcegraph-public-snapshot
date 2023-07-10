package compute

type Text struct {
	Value string `json:"value"`
	Kind  string `json:"kind"`
}

// TextExtra provides extra contextual information on top of the Text result.
type TextExtra struct {
	Text
	RepositoryID int32  `json:"repositoryID"`
	Repository   string `json:"repository"`
}
