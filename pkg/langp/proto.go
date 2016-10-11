package langp

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

// Error is returned in the event of any request error, in addition to the HTTP
// status 400 Bad Request.
type Error struct {
	// ErrorMsg, if any, specifies that there was an error serving the request.
	ErrorMsg string `json:"Error"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.ErrorMsg
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

// LSP converts this langp position into its closest LSP equivalent.
func (p Position) LSP() *lsp.TextDocumentPositionParams {
	return &lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: p.File},
		Position: lsp.Position{
			Line:      p.Line,
			Character: p.Character,
		},
	}
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

// LSP converts this langp range into its LSP equivalent.
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

// RepoRev represents a repository at a specific commit.
type RepoRev struct {
	// Repo is the repository URI.
	Repo string

	// Commit is the Git commit ID (not branch) of the repository.
	Commit string
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

// Symbol is a symbol in code.
type Symbol struct {
	// DefSpec is the DefSpec for this symbol.
	DefSpec

	// Name of the symbol. This need not be unique.
	Name string

	// Kind is the kind of thing this definition is. This is
	// language-specific. Possible values include "type", "func",
	// "var", etc.
	Kind string

	// File is the path to the file containing the symbol.
	File string

	// DocHTML is the docstring for the symbol, in the format 'text/html'.
	//
	// Note: You can't assume DocHTML has already been sanitized.
	DocHTML string
}

// RefLocations contains a list of locations of references to a specific definition.
type RefLocations struct {
	// Refs is a list of references to a definition defined within the requested
	// repository.
	Refs []*Range
}

// ExternalRefs contains a list of all Defs used in a repository, but defined
// outside of it.
type ExternalRefs struct {
	Refs []*Ref
}

// Ref represents a reference to a definition.
type Ref struct {
	// Def is the definition that is being referenced. Because refs are always
	// global (i.e. use the default VCS branch), Def.Commit should always be an
	// empty string.
	Def *DefSpec

	// File is the file in which the reference to Def is made.
	File string

	// Line is the line in the file at which the reference to Def is made.
	Line int

	// Column is the line column at which the reference to Def is made.
	Column int
}

// Symbols contains a list of Defs available within a repository.
type Symbols struct {
	Symbols []*lsp.SymbolInformation
}

// ExportedSymbols contains a list of all Defs available for use by other
// repositories.
type ExportedSymbols struct {
	Symbols []*Symbol
}

// SymbolsQuery is a request for a set of symbols within a repo.
type SymbolsQuery struct {
	RepoRev

	// Query specifies the desired options for filtering the available symbols.
	Query string
}

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

// File is returned by ResolveFile to convert workspace paths into objects
// Sourcegraph understands.
type File struct {
	// Repo is the repository URI in which the file is located.
	Repo string

	// Commit is the Git commit ID (not branch) of the repository.
	Commit string

	// Path is the file which the user is viewing, relative to the repository root.
	Path string
}
