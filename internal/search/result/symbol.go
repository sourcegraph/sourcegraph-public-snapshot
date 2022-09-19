package result

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcegraph/go-lsp"
)

// Symbol is a code symbol.
type Symbol struct {
	Name string

	// TODO (@camdencheek): remove path since it's duplicated
	// in the file reference of symbol match. Alternatively,
	// merge Symbol and SymbolMatch.
	Path       string
	Line       int
	Character  int
	Kind       string
	Language   string
	Parent     string
	ParentKind string
	Signature  string

	FileLimited bool
}

// NewSymbolMatch returns a new SymbolMatch. Passing -1 as the character will make NewSymbolMatch infer
// the column from the line and symbol name.
func NewSymbolMatch(file *File, lineNumber, character int, name, kind, parent, parentKind, language, line string, fileLimited bool) *SymbolMatch {
	if character == -1 {
		// The caller is requesting we infer the character position.
		character = strings.Index(line, name)

		if character == -1 {
			// We couldn't find the symbol in the line, so set the column to 0. ctags doesn't always
			// return the right line.
			character = 0
		}
	}

	return &SymbolMatch{
		Symbol: Symbol{
			Name:        name,
			Kind:        kind,
			Parent:      parent,
			ParentKind:  parentKind,
			Path:        file.Path,
			Line:        lineNumber,
			Character:   character,
			Language:    language,
			FileLimited: fileLimited,
		},
		File: file,
	}
}

func (s Symbol) NormalizedKind() string {
	if s.Kind != "" {
		return s.Kind
	}
	return "unknown"
}

func (s Symbol) Range() lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: s.Line - 1, Character: s.Character},
		End:   lsp.Position{Line: s.Line - 1, Character: s.Character + len(s.Name)},
	}
}

// Symbols is the result of a search on the symbols service.
type Symbols = []Symbol

// SymbolMatch is a symbol search result decorated with extra metadata in the frontend.
type SymbolMatch struct {
	Symbol Symbol
	File   *File
}

func (s *SymbolMatch) URL() *url.URL {
	base := s.File.URL()
	base.RawQuery = urlFragmentFromRange(s.Symbol.Range())
	return base
}

func urlFragmentFromRange(lspRange lsp.Range) string {
	if lspRange.Start == lspRange.End {
		return "L" + lineSpecFromPosition(lspRange.Start, false)
	}

	hasCharacter := lspRange.Start.Character != 0 || lspRange.End.Character != 0
	return "L" + lineSpecFromPosition(lspRange.Start, hasCharacter) + "-" + lineSpecFromPosition(lspRange.End, hasCharacter)
}

func lineSpecFromPosition(pos lsp.Position, forceIncludeCharacter bool) string {
	if !forceIncludeCharacter && pos.Character == 0 {
		return strconv.Itoa(pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", pos.Line+1, pos.Character+1)
}

func pick(symbols []*SymbolMatch, satisfy func(*SymbolMatch) bool) []*SymbolMatch {
	var result []*SymbolMatch
	for _, symbol := range symbols {
		if satisfy(symbol) {
			result = append(result, symbol)
		}
	}
	return result
}

func SelectSymbolKind(symbols []*SymbolMatch, field string) []*SymbolMatch {
	return pick(symbols, func(s *SymbolMatch) bool {
		return field == s.Symbol.Kind
	})
}
