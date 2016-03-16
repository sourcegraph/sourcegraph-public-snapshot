package payloads

import (
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
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
}

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
	Contents  string
	StartLine uint32
	EndLine   uint32
}
