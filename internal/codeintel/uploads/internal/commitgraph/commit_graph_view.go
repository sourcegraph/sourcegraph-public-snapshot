package commitgraph

// CommitGraphView is a space-efficient view of a commit graph decorated with the
// set of uploads visible at every commit.
type CommitGraphView struct {
	// Meta is a map from commit to metadata on each visible upload from that
	// commit's location in the commit graph.
	Meta map[string][]UploadMeta

	// Tokens is a map from upload identifiers to a hash of their root an indexer
	// field. Equality of this token for two uploads means that they are able to
	// "shadow" one another.
	Tokens map[int]string
}

// UploadMeta represents the visibility of an LSIF upload from a particular location
// on a repository's commit graph.
type UploadMeta struct {
	UploadID int
	Distance uint32
}

func NewCommitGraphView() *CommitGraphView {
	return &CommitGraphView{
		Meta:   map[string][]UploadMeta{},
		Tokens: map[int]string{},
	}
}

func (v *CommitGraphView) Add(meta UploadMeta, commit, token string) {
	v.Meta[commit] = append(v.Meta[commit], meta)
	v.Tokens[meta.UploadID] = token
}
