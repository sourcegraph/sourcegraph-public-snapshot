package result

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/gituri"
)

// Symbol is a code symbol.
type Symbol struct {
	Name       string
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

// Symbols is the result of a search on the symbols service.
type Symbols = []Symbol

// SymbolMatch is a symbol search result decorated with extra metadata in the frontend.
type SymbolMatch struct {
	Symbol  Symbol
	BaseURI *gituri.URI
}

func (s *SymbolMatch) URI() *gituri.URI {
	return s.BaseURI.WithFilePath(s.Symbol.Path)
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
