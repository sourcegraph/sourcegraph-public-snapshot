package payloads

import (
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
)

// DefCommon holds all of the def-specific information necessary to
// render a def page template. It is returned by GetDefCommon. It is assumes
// that pages rendered are also provided with repoCommon,
// repoRevCommon, and repoBuildCommon template data.
type DefCommon struct {
	// Def holds information about this definition.
	Def *sourcegraph.Def `json:"Data,omitempty"`

	// QualifiedName is the user-friendly form of this definition containing its
	// name and type, and is used as the title for this definition.
	QualifiedName *pbtypes.HTML

	// URL is the DefKey-based URL that can be used to request this definition.
	URL string

	// File holds information about the file that contains the declaration of
	// this definition.
	File sourcegraph.TreeEntrySpec

	// ByteStartPosition and ByteEndPosition are the byte offsets that this
	// definition occupy in the original containing file.
	ByteStartPosition, ByteEndPosition uint32

	// Found, when false, indicates that a definition does indeed exist, but
	// has not yet been indexed by our build system.
	Found bool
}

// Changeset holds the payload that will be sent when requesting information about
// changesets.
type Changeset struct {
	// Changeset contains the changeset data.
	Changeset *sourcegraph.Changeset

	// Delta contains information about the difference between the revisions
	// in the changeset.
	Delta *sourcegraph.Delta

	// BaseTip is the tip commit of the `base` of this changeset. This is compared
	// against the Delta's BaseCommit which is the merge-base commit. If they
	// differ, the changeset can not be cleanly merged.
	BaseTip *vcs.Commit

	// Commits holds the list of commits that are part of this changeset, augmented
	// with user information.
	Commits []*AugmentedCommit `json:",omitempty"`

	// Files holds the actual changeset files (diffs).
	Files *sourcegraph.DeltaFiles `json:",omitempty"`

	// Reviews holds all the reviews that were made on this changeset.
	Reviews []*sourcegraph.ChangesetReview

	// Events is a list of events that occurred on this changeset. Events are
	// popuplated when a changeset is updated.
	Events []*sourcegraph.ChangesetEvent
}

// CodeFile holds information about a code file to be displayed in the UI.
type CodeFile struct {
	// Repo is the repository that the file belongs to.
	Repo *sourcegraph.Repo

	// RepoCommit is the commit that the file belongs to.
	RepoCommit *AugmentedCommit

	// RepoBuildInfo contains information about the build status of this commit.
	RepoBuildInfo *sourcegraph.RepoBuildInfo

	// Entry is the actual file data.
	Entry *sourcegraph.TreeEntry

	// EntrySpec contains information about the file, such as the path and
	// revision.
	EntrySpec sourcegraph.TreeEntrySpec
}

// AugmentedCommit is an augmented commit for presentation in the app. It is
// displayed with the Commit partial template in commits.inc.html.
type AugmentedCommit struct {
	*vcs.Commit
	AuthorPerson, CommitterPerson *sourcegraph.Person
	RepoURI                       string
}

type TokenSearchResult struct {
	// Def holds information about this definition.
	Def *sourcegraph.Def

	// QualifiedName is the user-friendly form of this definition containing its
	// name and type, and is used as the title for this definition.
	QualifiedName *pbtypes.HTML

	// URL is the DefKey-based URL that can be used to request this definition.
	URL string
}

type TextSearchResult struct {
	File      string
	Lines     []*pbtypes.HTML
	StartLine uint32
	EndLine   uint32
}
