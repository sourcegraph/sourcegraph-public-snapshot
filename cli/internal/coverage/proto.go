package coverage

import (
	"fmt"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

// Hover represents a message for when a user "hovers" over a definition. It is
// a human-readable description of a definition.
type Hover struct {
	// Title is a human-readable string of the
	// definition. Typically this is the type/function signature.
	Title string

	// DocHTML is the docstring for the definition, in the format
	// 'text/html'.
	//
	// Note: You can't assume DocHTML has already been sanitized.
	DocHTML string

	// DefSpec is optionally set, and is the DefSpec for the hover.
	DefSpec *DefSpec `json:",omitempty"`

	// Unresolved is optionally set, and if true indicates that there is
	// no hover content for this position (eg a comment)
	Unresolved bool
}

func HoverFromLSP(l *lsp.Hover) *Hover {
	if len(l.Contents) == 0 {
		return &Hover{}
	}
	var docHTML string
	for _, m := range l.Contents[1:] {
		if m.Language == "text/html" {
			docHTML = m.Value
		}
	}
	return &Hover{
		Title:   l.Contents[0].Value,
		DocHTML: docHTML,
	}
}

// RefLocations contains a list of locations of references to a specific definition.
type RefLocations struct {
	// Refs is a list of references to a definition defined within the requested
	// repository.
	Refs []*Range
}

// Range represents a specific range within a file.
type Range struct {
	// Repo is the repository URI in which the file is located.
	Repo string

	// Commit is the Git commit ID (not branch) of the repository.
	Commit string

	// File is the file which the user is viewing, relative to the repository root.
	File string

	// StartLine is the starting line number in the file (zero based), i.e.
	// where the range starts.
	StartLine int

	// EndLine is the ending line number in the file (zero based), i.e. where
	// the range ends.
	EndLine int

	// StartCharacter is the starting character offset on the starting line in
	// the file (zero based).
	StartCharacter int

	// EndCharacter is the ending character offset on the ending line in the
	// file (zero based).
	EndCharacter int
}

// LSP converts this range into its LSP equivalent.
func (r Range) LSP() lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      r.StartLine,
			Character: r.StartCharacter,
		},
		End: lsp.Position{
			Line:      r.EndLine,
			Character: r.EndCharacter,
		},
	}
}

// Empty returns true if the range has no fields set.
func (r Range) Empty() bool {
	return r == Range{}
}

// DefSpec is a globally unique identifier for a definition in a repository at
// a specific revision. It is the same as the Srclib DefSpec.
type DefSpec struct {
	// Repo is the repository URI (example github.com/gorilla/mux).
	Repo string

	// Commit is the Git commit ID (not branch) of the repository.
	Commit string `json:",omitempty"`

	// UnitType (example GoPackage).
	UnitType string

	// Unit (example net/http, or github.com/gorilla/mux/subpkg).
	Unit string

	// Path (example NewRequest).
	Path string
}

// String returns complete string representation of DefSpec.
func (d DefSpec) String() string {
	var rev string
	if d.Commit != "" {
		rev = "@" + d.Commit
	}
	return fmt.Sprintf("%s/%s%s/-/%s", d.UnitType, d.Unit, rev, d.Path)
}

// DefString returns partial string representation without revision of DefSpec.
func (d DefSpec) DefString() string {
	return fmt.Sprintf("%s/%s/-/%s", d.UnitType, d.Unit, d.Path)
}

// Position represents a single specific position within a file located in a
// repository at a given revision.
type Position struct {
	// Repo is the repository URI in which the file is located.
	Repo string

	// Commit is the Git commit ID (not branch) of the repository.
	Commit string

	// File is the file which the user is viewing, relative to the repository root.
	File string

	// Line is the line number in the file (zero based), e.g. where a user's cursor
	// is located within the file.
	Line int

	// Character is the character offset on a line in the file (zero based), e.g.
	// where a user's cursor is located within the file.
	Character int
}

// LSP converts this position into its closest LSP equivalent.
func (p Position) LSP() *lsp.TextDocumentPositionParams {
	return &lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: p.File},
		Position: lsp.Position{
			Line:      p.Line,
			Character: p.Character,
		},
	}
}
