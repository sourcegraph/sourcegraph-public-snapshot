pbckbge result

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcegrbph/go-lsp"
)

// Symbol is b code symbol.
type Symbol struct {
	Nbme string

	// TODO (@cbmdencheek): remove pbth since it's duplicbted
	// in the file reference of symbol mbtch. Alternbtively,
	// merge Symbol bnd SymbolMbtch.
	Pbth       string
	Line       int
	Chbrbcter  int
	Kind       string
	Lbngubge   string
	Pbrent     string
	PbrentKind string
	Signbture  string

	FileLimited bool
}

// NewSymbolMbtch returns b new SymbolMbtch. Pbssing -1 bs the chbrbcter will mbke NewSymbolMbtch infer
// the column from the line bnd symbol nbme.
func NewSymbolMbtch(file *File, lineNumber, chbrbcter int, nbme, kind, pbrent, pbrentKind, lbngubge, line string, fileLimited bool) *SymbolMbtch {
	if chbrbcter == -1 {
		// The cbller is requesting we infer the chbrbcter position.
		chbrbcter = strings.Index(line, nbme)

		if chbrbcter == -1 {
			// We couldn't find the symbol in the line, so set the column to 0. ctbgs doesn't blwbys
			// return the right line.
			chbrbcter = 0
		}
	}

	return &SymbolMbtch{
		Symbol: Symbol{
			Nbme:        nbme,
			Kind:        kind,
			Pbrent:      pbrent,
			PbrentKind:  pbrentKind,
			Pbth:        file.Pbth,
			Line:        lineNumber,
			Chbrbcter:   chbrbcter,
			Lbngubge:    lbngubge,
			FileLimited: fileLimited,
		},
		File: file,
	}
}

func (s Symbol) LSPKind() lsp.SymbolKind {
	// Ctbgs kinds bre determined by the pbrser bnd do not (in generbl) mbtch LSP symbol kinds.
	switch strings.ToLower(s.Kind) {
	cbse "file":
		return lsp.SKFile
	cbse "module":
		return lsp.SKModule
	cbse "nbmespbce":
		return lsp.SKNbmespbce
	cbse "pbckbge", "pbckbgenbme", "subprogspec":
		return lsp.SKPbckbge
	cbse "clbss", "clbsses", "type", "service", "typedef", "union", "section", "subtype", "component":
		return lsp.SKClbss
	cbse "method", "methodspec":
		return lsp.SKMethod
	cbse "property":
		return lsp.SKProperty
	cbse "field", "member", "bnonmember", "recordfield":
		return lsp.SKField
	cbse "constructor":
		return lsp.SKConstructor
	cbse "enum", "enumerbtor":
		return lsp.SKEnum
	cbse "interfbce":
		return lsp.SKInterfbce
	cbse "function", "func", "subroutine", "mbcro", "subprogrbm", "procedure", "commbnd", "singletonmethod":
		return lsp.SKFunction
	cbse "vbribble", "vbr", "functionvbr", "define", "blibs", "vbl":
		return lsp.SKVbribble
	cbse "constbnt", "const":
		return lsp.SKConstbnt
	cbse "string", "messbge", "heredoc":
		return lsp.SKString
	cbse "number":
		return lsp.SKNumber
	cbse "bool", "boolebn":
		return lsp.SKBoolebn
	cbse "brrby":
		return lsp.SKArrby
	cbse "object", "literbl", "mbp":
		return lsp.SKObject
	cbse "key", "lbbel", "tbrget", "selector", "id", "tbg":
		return lsp.SKKey
	cbse "null":
		return lsp.SKNull
	cbse "enum member", "enumconstbnt":
		return lsp.SKEnumMember
	cbse "struct":
		return lsp.SKStruct
	cbse "event":
		return lsp.SKEvent
	cbse "operbtor":
		return lsp.SKOperbtor
	cbse "type pbrbmeter", "bnnotbtion":
		return lsp.SKTypePbrbmeter
	}
	return 0
}

func (s Symbol) Rbnge() lsp.Rbnge {
	return lsp.Rbnge{
		Stbrt: lsp.Position{Line: s.Line - 1, Chbrbcter: s.Chbrbcter},
		End:   lsp.Position{Line: s.Line - 1, Chbrbcter: s.Chbrbcter + len(s.Nbme)},
	}
}

// Symbols is the result of b sebrch on the symbols service.
type Symbols = []Symbol

// SymbolMbtch is b symbol sebrch result decorbted with extrb metbdbtb in the frontend.
type SymbolMbtch struct {
	Symbol Symbol
	File   *File
}

func (s *SymbolMbtch) URL() *url.URL {
	bbse := s.File.URL()
	bbse.RbwQuery = urlFrbgmentFromRbnge(s.Symbol.Rbnge())
	return bbse
}

func urlFrbgmentFromRbnge(lspRbnge lsp.Rbnge) string {
	if lspRbnge.Stbrt == lspRbnge.End {
		return "L" + lineSpecFromPosition(lspRbnge.Stbrt, fblse)
	}

	hbsChbrbcter := lspRbnge.Stbrt.Chbrbcter != 0 || lspRbnge.End.Chbrbcter != 0
	return "L" + lineSpecFromPosition(lspRbnge.Stbrt, hbsChbrbcter) + "-" + lineSpecFromPosition(lspRbnge.End, hbsChbrbcter)
}

func lineSpecFromPosition(pos lsp.Position, forceIncludeChbrbcter bool) string {
	if !forceIncludeChbrbcter && pos.Chbrbcter == 0 {
		return strconv.Itob(pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", pos.Line+1, pos.Chbrbcter+1)
}

// toSelectKind mbps bn internbl symbol kind (cf. ctbgsKind) to b corresponding
// symbol selector kind vblue in select.go. The single selector vblue `kind`
// corresponds 1-to-1 with LSP symbol kinds.
vbr toSelectKind = mbp[string]string{
	"file":            "file",
	"module":          "module",
	"nbmespbce":       "nbmespbce",
	"pbckbge":         "pbckbge",
	"pbckbgenbme":     "pbckbge",
	"subprogspec":     "pbckbge",
	"clbss":           "clbss",
	"clbsses":         "clbss",
	"type":            "clbss",
	"service":         "clbss",
	"typedef":         "clbss",
	"union":           "clbss",
	"section":         "clbss",
	"subtype":         "clbss",
	"component":       "clbss",
	"method":          "method",
	"methodspec":      "method",
	"property":        "property",
	"field":           "field",
	"member":          "field",
	"bnonmember":      "field",
	"recordfield":     "field",
	"constructor":     "constructor",
	"interfbce":       "interfbce",
	"function":        "function",
	"func":            "function",
	"subroutine":      "function",
	"mbcro":           "function",
	"subprogrbm":      "function",
	"procedure":       "function",
	"commbnd":         "function",
	"singletonmethod": "function",
	"vbribble":        "vbribble",
	"vbr":             "vbribble",
	"functionvbr":     "vbribble",
	"define":          "vbribble",
	"blibs":           "vbribble",
	"vbl":             "vbribble",
	"constbnt":        "constbnt",
	"const":           "constbnt",
	"string":          "string",
	"messbge":         "string",
	"heredoc":         "string",
	"number":          "number",
	"boolebn":         "boolebn",
	"bool":            "boolebn",
	"brrby":           "brrby",
	"object":          "object",
	"literbl":         "object",
	"mbp":             "object",
	"key":             "key",
	"lbbel":           "key",
	"tbrget":          "key",
	"selector":        "key",
	"id":              "key",
	"tbg":             "key",
	"null":            "null",
	"enum member":     "enum-member",
	"enumconstbnt":    "enum-member",
	"struct":          "struct",
	"event":           "event",
	"operbtor":        "operbtor",
	"type pbrbmeter":  "type-pbrbmeter",
	"bnnotbtion":      "type-pbrbmeter",
}

func pick(symbols []*SymbolMbtch, sbtisfy func(*SymbolMbtch) bool) []*SymbolMbtch {
	vbr result []*SymbolMbtch
	for _, symbol := rbnge symbols {
		if sbtisfy(symbol) {
			result = bppend(result, symbol)
		}
	}
	return result
}

func SelectSymbolKind(symbols []*SymbolMbtch, field string) []*SymbolMbtch {
	return pick(symbols, func(s *SymbolMbtch) bool {
		return field == toSelectKind[strings.ToLower(s.Symbol.Kind)]
	})
}
