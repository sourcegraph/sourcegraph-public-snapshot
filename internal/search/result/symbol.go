package result

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

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
	Kind       string
	Language   string
	Parent     string
	ParentKind string
	Signature  string
	Pattern    string

	FileLimited bool
}

// offset calculates a symbol offset based on the the only Symbol
// data member that currently exposes line content: the symbols Pattern member,
// which has the form /^ ... $/. We find the offset of the symbol name in this
// line, after escaping the Pattern.
func (s *Symbol) offset() int {
	if s.Pattern == "" {
		return 0
	}
	i := strings.Index(unescapePattern(s.Pattern), s.Name)
	if i >= 0 {
		return i
	}
	return 0
}

// unescapePattern expects a regexp pattern of the form /^ ... $/ and unescapes
// the pattern inside it.
func unescapePattern(pattern string) string {
	pattern = strings.TrimSuffix(strings.TrimPrefix(pattern, "/^"), "$/")
	var start int
	var r rune
	var escaped []rune
	buf := []byte(pattern)

	next := func() rune {
		r, start := utf8.DecodeRune(buf)
		buf = buf[start:]
		return r
	}

	for len(buf) > 0 {
		r = next()
		if r == '\\' && len(buf[start:]) > 0 {
			r = next()
			if r == '/' || r == '\\' {
				escaped = append(escaped, r)
				continue
			}
			escaped = append(escaped, '\\', r)
			continue
		}
		escaped = append(escaped, r)
	}
	return string(escaped)
}

func (s Symbol) Range() lsp.Range {
	offset := s.offset()
	return lsp.Range{
		Start: lsp.Position{Line: s.Line - 1, Character: offset},
		End:   lsp.Position{Line: s.Line - 1, Character: offset + len(s.Name)},
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
	base.Fragment = urlFragmentFromRange(s.Symbol.Range())
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

// toSelectKind maps an internal symbol kind (cf. ctagsKind) to a corresponding
// symbol selector kind value in select.go. The single selector value `kind`
// corresponds 1-to-1 with LSP symbol kinds.
var toSelectKind = map[string]string{
	"file":            "file",
	"module":          "module",
	"namespace":       "namespace",
	"package":         "package",
	"packagename":     "package",
	"subprogspec":     "package",
	"class":           "class",
	"type":            "class",
	"service":         "class",
	"typedef":         "class",
	"union":           "class",
	"section":         "class",
	"subtype":         "class",
	"component":       "class",
	"method":          "method",
	"methodspec":      "method",
	"property":        "property",
	"field":           "field",
	"member":          "field",
	"anonmember":      "field",
	"recordfield":     "field",
	"constructor":     "constructor",
	"interface":       "interface",
	"function":        "function",
	"func":            "function",
	"subroutine":      "function",
	"macro":           "function",
	"subprogram":      "function",
	"procedure":       "function",
	"command":         "function",
	"singletonmethod": "function",
	"variable":        "variable",
	"var":             "variable",
	"functionvar":     "variable",
	"define":          "variable",
	"alias":           "variable",
	"val":             "variable",
	"constant":        "constant",
	"const":           "constant",
	"string":          "string",
	"message":         "string",
	"heredoc":         "string",
	"number":          "number",
	"boolean":         "boolean",
	"bool":            "boolean",
	"array":           "array",
	"object":          "object",
	"literal":         "object",
	"map":             "object",
	"key":             "key",
	"label":           "key",
	"target":          "key",
	"selector":        "key",
	"id":              "key",
	"tag":             "key",
	"null":            "null",
	"enum member":     "enum-member",
	"enumconstant":    "enum-member",
	"struct":          "struct",
	"event":           "event",
	"operator":        "operator",
	"type parameter":  "type-parameter",
	"annotation":      "type-parameter",
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
		return field == toSelectKind[strings.ToLower(s.Symbol.Kind)]
	})
}
