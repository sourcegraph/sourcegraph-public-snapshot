package dbstore

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

// UploadMeta represents the visibility of an LSIF upload from a particular location
// on a repository's commit graph. The Flags field describes the visibility of the
// upload from the current viewer's perspective.
type UploadMeta struct {
	UploadID int

	// Flags encodes the flags FlagAncestorVisible and FlagOverwritten and leaves
	// the remaining lower 30-bits to encode an unsigned distance between commits.
	Flags uint32
}

const (
	// FlagAncestorVisible indicates whether or not this upload was visible to the commit
	// by looking for older uploads defined in an ancestor commit. False indicates that
	// the upload is only visible via another upload defined in a descendant commit.
	FlagAncestorVisible uint32 = (1 << 30)

	// FlagOverwritten indicates that this upload defined on an ancestor commit that is
	// farther away than an upload defined on a descendant commit that has the same root
	// and indexer.
	//
	// If overwritten, this upload is not considered in nearest commit operations, but
	// is kept in the database so that we can reconstruct the set of all ancestor-visible
	// uploads of a commit, which is useful when determining the closest uploads with only
	// a partial commit graph.
	FlagOverwritten uint32 = (1 << 29)

	// MaxDistance is the maximum encodeable distance between two commits.
	//
	// This value can be used to quickly remove flags from the encoded distance value.
	MaxDistance = ^(FlagAncestorVisible | FlagOverwritten)
)
