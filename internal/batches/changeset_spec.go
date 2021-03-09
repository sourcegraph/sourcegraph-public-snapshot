package batches

// These struct definitions must line up with the changeset spec schema, as
// defined in sourcegraph/sourcegraph.

type ChangesetSpec struct {
	BaseRepository string `json:"baseRepository"`

	*ExternalChangeset
	*CreatedChangeset
}

type ExternalChangeset struct {
	ExternalID string `json:"externalID"`
}

type CreatedChangeset struct {
	BaseRef        string                 `json:"baseRef"`
	BaseRev        string                 `json:"baseRev"`
	HeadRepository string                 `json:"headRepository"`
	HeadRef        string                 `json:"headRef"`
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	Commits        []GitCommitDescription `json:"commits"`
	Published      interface{}            `json:"published"`
}

type GitCommitDescription struct {
	Message     string `json:"message"`
	Diff        string `json:"diff"`
	AuthorName  string `json:"authorName,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
}
