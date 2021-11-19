package symbols

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// symbolInDB is the same as `protocol.Symbol`, but with two additional columns:
// namelowercase and pathlowercase, which enable indexed case insensitive
// queries.
type symbolInDB struct {
	Name          string
	NameLowercase string // derived from `Name`
	Path          string
	PathLowercase string // derived from `Path`
	Line          int
	Kind          string
	Language      string
	Parent        string
	ParentKind    string
	Signature     string
	Pattern       string

	// Whether or not the symbol is local to the file.
	FileLimited bool
}

func symbolToSymbolInDB(symbol result.Symbol) symbolInDB {
	return symbolInDB{
		Name:          symbol.Name,
		NameLowercase: strings.ToLower(symbol.Name),
		Path:          symbol.Path,
		PathLowercase: strings.ToLower(symbol.Path),
		Line:          symbol.Line,
		Kind:          symbol.Kind,
		Language:      symbol.Language,
		Parent:        symbol.Parent,
		ParentKind:    symbol.ParentKind,
		Signature:     symbol.Signature,
		Pattern:       symbol.Pattern,

		FileLimited: symbol.FileLimited,
	}
}

func symbolInDBToSymbol(symbolInDB symbolInDB) result.Symbol {
	return result.Symbol{
		Name:       symbolInDB.Name,
		Path:       symbolInDB.Path,
		Line:       symbolInDB.Line,
		Kind:       symbolInDB.Kind,
		Language:   symbolInDB.Language,
		Parent:     symbolInDB.Parent,
		ParentKind: symbolInDB.ParentKind,
		Signature:  symbolInDB.Signature,
		Pattern:    symbolInDB.Pattern,

		FileLimited: symbolInDB.FileLimited,
	}
}
