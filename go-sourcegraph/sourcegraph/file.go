package sourcegraph

import "html/template"

// Range is a character range within a file.
//
// TODO(sqs): rune range?
type Range struct {
	Start, End int
}

// FileData represents a range of (possibly annotated) data in a file.
type FileData struct {
	// Repo is the repository that contains this file data.
	RepoRev RepoRevSpec

	// File is the file (relative to the repository root directory) that this
	// file data is from.
	File string

	// Start and End is the range of the data in the file.
	Range *Range `json:",omitempty"`

	// EntireFile is true if the data spans the entire file contents.
	EntireFile bool `json:",omitempty"`

	// Annotated (i.e., HTML-marked up) content.
	Annotated template.HTML

	// Raw data.
	Raw []byte
}
