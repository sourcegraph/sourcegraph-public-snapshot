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

func NewSymbolMatch(file *File, lineNumber int, name, kind, parent, parentKind, language, line string, fileLimited bool) *SymbolMatch {
	return &SymbolMatch{
		Symbol: Symbol{
			Name:       name,
			Kind:       kind,
			Parent:     parent,
			ParentKind: parentKind,
			Path:       file.Path,
			Line:       lineNumber,
			Language:   language,
			// symbolRange requires a Pattern /^...$/ containing the line of the symbol to compute the symbol's offsets.
			// This Pattern is directly accessible on the unindexed code path. But on the indexed code path, we need to
			// populate it, or we will always compute a 0 offset, which messes up API use (e.g., highlighting).
			// It must escape `/` or `\` in the line.
			Pattern:     fmt.Sprintf("/^%s$/", escape(line)),
			FileLimited: fileLimited,
		},
		File: file,
	}
}

func (s Symbol) LSPKind() lsp.SymbolKind {
	// Ctags kinds are determined by the parser and do not (in general) match LSP symbol kinds.
	switch strings.ToLower(s.Kind) {
	case "file":
		return lsp.SKFile
	case "module":
		return lsp.SKModule
	case "namespace":
		return lsp.SKNamespace
	case "package", "packagename", "subprogspec":
		return lsp.SKPackage
	case "class", "classes", "type", "service", "typedef", "union", "section", "subtype", "component":
		return lsp.SKClass
	case "method", "methodspec":
		return lsp.SKMethod
	case "property":
		return lsp.SKProperty
	case "field", "member", "anonmember", "recordfield":
		return lsp.SKField
	case "constructor":
		return lsp.SKConstructor
	case "enum", "enumerator":
		return lsp.SKEnum
	case "interface":
		return lsp.SKInterface
	case "function", "func", "subroutine", "macro", "subprogram", "procedure", "command", "singletonmethod":
		return lsp.SKFunction
	case "variable", "var", "functionvar", "define", "alias", "val":
		return lsp.SKVariable
	case "constant", "const":
		return lsp.SKConstant
	case "string", "message", "heredoc":
		return lsp.SKString
	case "number":
		return lsp.SKNumber
	case "bool", "boolean":
		return lsp.SKBoolean
	case "array":
		return lsp.SKArray
	case "object", "literal", "map":
		return lsp.SKObject
	case "key", "label", "target", "selector", "id", "tag":
		return lsp.SKKey
	case "null":
		return lsp.SKNull
	case "enum member", "enumconstant":
		return lsp.SKEnumMember
	case "struct":
		return lsp.SKStruct
	case "event":
		return lsp.SKEvent
	case "operator":
		return lsp.SKOperator
	case "type parameter", "annotation":
		return lsp.SKTypeParameter
	}
	return 0
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

func escape(s string) string {
	isSpecial := func(c rune) bool {
		switch c {
		case '\\', '/':
			return true
		default:
			return false
		}
	}

	// Avoid extra work by counting additions. regexp.QuoteMeta does the same,
	// but is more efficient since it does it via bytes.
	count := 0
	for _, c := range s {
		if isSpecial(c) {
			count++
		}
	}
	if count == 0 {
		return s
	}

	escaped := make([]rune, 0, len(s)+count)
	for _, c := range s {
		if isSpecial(c) {
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, c)
	}
	return string(escaped)
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
	"classes":         "class",
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
