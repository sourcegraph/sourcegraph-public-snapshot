package sourcegraph

import "html/template"

// Range is a byte range within a file.
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

	// Range (of bytes) of the data in the file.
	Range *Range `json:",omitempty"`

	// EntireFile is true if the data spans the entire file contents.
	EntireFile bool `json:",omitempty"`

	// Annotated (i.e., HTML-marked up) content.
	Annotated template.HTML

	// Raw data.
	Raw []byte
}

// FileWithRange is returned by GetFileWithOptions and includes the
// returned file's BasicTreeEntry as well as the actual range of lines and
// bytes returned (based on the GetFileOptions parameters). That is,
// if Start/EndLine are set in GetFileOptions, this struct's
// Start/EndByte will be set to the actual start and end bytes of
// those specified lines, and so on for the other fields in
// GetFileOptions.
type FileWithRange struct {
	*BasicTreeEntry
	FileRange // range of actual returned tree entry contents within file
}
