package payloads

import (
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// CodeFile holds information about a code file to be displayed in the UI.
type CodeFile struct {
	// Repo is the repository that the file belongs to.
	Repo *sourcegraph.Repo

	// RepoCommit is the commit that the file belongs to.
	RepoCommit *AugmentedCommit

	// SrclibDataVersion contains information about this file's srclib
	// analysis.
	SrclibDataVersion *sourcegraph.SrclibDataVersion

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
