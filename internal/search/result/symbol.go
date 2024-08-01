package result

import (
	"cmp"
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

func (s Symbol) LSPKind() lsp.SymbolKind {
	kind, ok := lspKinds[s.Kind]
	if !ok {
		return 0
	}
	return kind
}

// SelectKind maps an internal symbol kind (cf. ctagsKind) to a corresponding
// symbol selector kind value in select.go. The single selector value `kind`
// corresponds 1-to-1 with LSP symbol kinds.
func (s Symbol) SelectKind() (string, bool) {
	kind, ok := lspKinds[s.Kind]
	if !ok {
		return "", false
	}

	kindName := strings.ToLower(kind.String())
	switch kindName {
	case "enummember":
		return "enum-member", true
	case "typeparameter":
		return "type-parameter", true
	default:
		return kindName, true
	}
}

func (s Symbol) Range() lsp.Range {
	// TODO(keegancsmith) For results from zoekt s.Character is not the start
	// of the symbol, but the start of the match. So doing s.Character +
	// len(s.Name) is incorrect.
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

func compareSymbolMatches(a, b *SymbolMatch) int {
	if v := cmp.Compare(a.Symbol.Line, b.Symbol.Line); v != 0 {
		return v
	}

	if v := cmp.Compare(a.Symbol.Name, b.Symbol.Name); v != 0 {
		return v
	}

	if v := cmp.Compare(a.Symbol.Kind, b.Symbol.Kind); v != 0 {
		return v
	}

	if v := cmp.Compare(a.Symbol.Parent, b.Symbol.Parent); v != 0 {
		return v
	}

	return cmp.Compare(a.Symbol.ParentKind, b.Symbol.ParentKind)
}

// DedupSymbols removes duplicate symbols from the list. We use a heuristic to
// determine duplicate matches. We regard a match as the same if they have the
// same symbol info and appear on the same line. I am unaware of a language
// where you can break that assumption.
func DedupSymbols(symbols []*SymbolMatch) []*SymbolMatch {
	if len(symbols) <= 1 {
		return symbols
	}
	dedup := symbols[:1]
	for _, sym := range symbols[1:] {
		last := dedup[len(dedup)-1]
		if compareSymbolMatches(sym, last) == 0 {
			continue
		}
		dedup = append(dedup, sym)
	}
	return dedup
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
		kind, ok := s.Symbol.SelectKind()
		return ok && field == kind
	})
}

// lspKinds maps a ctags kind to an LSP symbol kind. Ctags kinds are determined
// by the parser and do not (in general) match LSP symbol kinds.
var lspKinds = map[string]lsp.SymbolKind{
	"file":            lsp.SKFile,
	"module":          lsp.SKModule,
	"namespace":       lsp.SKNamespace,
	"package":         lsp.SKPackage,
	"packagename":     lsp.SKPackage,
	"subprogspec":     lsp.SKPackage,
	"class":           lsp.SKClass,
	"classes":         lsp.SKClass,
	"type":            lsp.SKClass,
	"service":         lsp.SKClass,
	"typedef":         lsp.SKClass,
	"union":           lsp.SKClass,
	"section":         lsp.SKClass,
	"subtype":         lsp.SKClass,
	"component":       lsp.SKClass,
	"accessor":        lsp.SKMethod,
	"getter":          lsp.SKMethod,
	"method":          lsp.SKMethod,
	"methodalias":     lsp.SKMethod,
	"methodspec":      lsp.SKMethod,
	"setter":          lsp.SKMethod,
	"singletonmethod": lsp.SKFunction,
	"property":        lsp.SKProperty,
	"field":           lsp.SKField,
	"member":          lsp.SKField,
	"anonmember":      lsp.SKField,
	"recordfield":     lsp.SKField,
	"constructor":     lsp.SKConstructor,
	"enum":            lsp.SKEnum,
	"enumerator":      lsp.SKEnum,
	"interface":       lsp.SKInterface,
	"function":        lsp.SKFunction,
	"func":            lsp.SKFunction,
	"subroutine":      lsp.SKFunction,
	"macro":           lsp.SKFunction,
	"subprogram":      lsp.SKFunction,
	"procedure":       lsp.SKFunction,
	"command":         lsp.SKFunction,
	"variable":        lsp.SKVariable,
	"var":             lsp.SKVariable,
	"functionvar":     lsp.SKVariable,
	"define":          lsp.SKVariable,
	"alias":           lsp.SKVariable,
	"val":             lsp.SKVariable,
	"constant":        lsp.SKConstant,
	"const":           lsp.SKConstant,
	"string":          lsp.SKString,
	"message":         lsp.SKString,
	"heredoc":         lsp.SKString,
	"number":          lsp.SKNumber,
	"bool":            lsp.SKBoolean,
	"boolean":         lsp.SKBoolean,
	"array":           lsp.SKArray,
	"object":          lsp.SKObject,
	"literal":         lsp.SKObject,
	"map":             lsp.SKObject,
	"key":             lsp.SKKey,
	"label":           lsp.SKKey,
	"target":          lsp.SKKey,
	"selector":        lsp.SKKey,
	"id":              lsp.SKKey,
	"tag":             lsp.SKKey,
	"null":            lsp.SKNull,
	"enum member":     lsp.SKEnumMember,
	"enumconstant":    lsp.SKEnumMember,
	"enummember":      lsp.SKEnumMember,
	"struct":          lsp.SKStruct,
	"event":           lsp.SKEvent,
	"operator":        lsp.SKOperator,
	"type parameter":  lsp.SKTypeParameter,
	"annotation":      lsp.SKTypeParameter,
}
