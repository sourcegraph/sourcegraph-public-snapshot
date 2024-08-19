// Package schema defines the common data format for elements of a Jupyter notebook.
//
// It is based on the [v4.4] definition, as it is stable and encompasses all the data
// necessary for accurate rendering. Note, that schema validation is not a goal of this
// package, and so, interfaces defined here will often omit the non-essential data,
// e.g. metadata or fields specific to JupyterLab environment. 
//
// [v4.4]: https://github.com/jupyter/nbformat/blob/main/nbformat/v4/nbformat.v4.4.schema.json
package schema

import (
	"fmt"
)

// Version specifies an [nbformat] version.
//
// [nbformat]: https://nbformat.readthedocs.io/en/latest/format_description.html
type Version struct {
	Major int
	Minor int
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

// Notebook is a common interface of the Jupyter Notebooks' different formats.
type Notebook interface {
	Version() Version
	Cells() []Cell
}

type NotebookMetadata interface {
	// Language reports the language of the document's associated kernel.
	Language() string
}

// Cell encapsulates the raw content of each notebook cell and its designated mime-type.
type Cell interface {
	Type() CellType
	MimeType() string
	Text() []byte
}

// HasAttachments is implemented by cells which include [cell attachments].
//
// [cell attachments]: https://nbformat.readthedocs.io/en/latest/format_description.html#cell-attachments
type HasAttachments interface {
	// Attachments are only defined for v4.0 and above for markdown and raw cells
	// and may be omitted in the JSON. Cells without attachments should return nil.
	Attachments() Attachments
}

// CellType reports the intended cell type to the components that work
// with notebook cells through the Cell interface.
//
// "markdown", "raw", "code", and "unrecognized" are official cell types,
// as defined in the [nbformat JSON schema]. Only "true" cells should use
// these to report their type.
//
// "execute_result", "display_data", "stream", and "error" are, on the other hand,
// [output types] and only appear within the code cells. Although they
// aren't "true" cells with respect to the Jupyter's schema, functionally they are
// quite similar (output structs will, in fact, implement the Cell interface).
// They should use the predefined values to report their .CellType().
//
// [nbformat JSON schema]: https://github.com/jupyter/nbformat/blob/dda25ef6565a33cef9096c283a47cc3fa8b96f91/nbformat/v4/nbformat.v4.5.schema.json#L108
// [output types]: https://github.com/jupyter/nbformat/blob/dda25ef6565a33cef9096c283a47cc3fa8b96f91/nbformat/v4/nbformat.v4.5.schema.json#L303
type CellType int

const (
	Unrecognized CellType = iota
	Markdown
	Raw
	Code

	ExecuteResult
	DisplayData
	Stream
	Error
)

// String returns the official enum values for cell and output types as defined in the JSON schema v4.5.
//
// These values should not be depended on, as their representation may change in the future.
// They are only provided to make logs and test output more readable.
func (t CellType) String() string {
	switch t {
	case Markdown:
		return "markdown"
	case Raw:
		return "raw"
	case Code:
		return "code"
	case ExecuteResult:
		return "execute_result"
	case DisplayData:
		return "display_data"
	case Stream:
		return "stream"
	case Error:
		return "error"
	}
	return "unrecognized"
}

// CodeCell contains the source code in the language of the document's associated kernel,
// and a list of outputs associated with executing that code.
type CodeCell interface {
	Cell
	Outputter
	ExecutionCounter

	Language() string
}

type ExecutionCounter interface {
	ExecutionCount() int
}

type Outputter interface {
	Outputs() []Cell
}

// MimeBundle holds rich display outputs, such as images, JSON, HTML, etc.
// A single output will often have 2 representations:
//  1. raw data, e.g. a base64-encoded image or JSON string/bytes, and
//  2. plain text representation of the original object (in Python it's the output of the __repr__() method).
//
// MimeBundle partially implements Cell interface, hiding the above complexity from the caller.
// When reporting MimeType implementations should prefer "text/html", "image/png", and any other type to "text/plain",
// and only return the latter if it is the only available option.
//
// Similarly, Text returns the value associated with the richer of the available mime-types.
type MimeBundle interface {
	MimeType() string
	Text() []byte

	// PlainText returns the value associated with "text/plain" mime-type if present and a nil slice otherwise.
	// A renderer may want to fallback to this option if it is not able to render the richer mime-type.
	PlainText() []byte
}

// Attachments are data for inline images stored as a mime-bundle keyed by filename.
type Attachments interface {
	// MimeBundle returns a mime-bundle associated with the filename.
	// If no data is present for the file, implementations should return nil.
	MimeBundle(filename string) MimeBundle
}
