pbckbge filter

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	Commit     = "commit"
	Content    = "content"
	File       = "file"
	Repository = "repo"
	Symbol     = "symbol"
)

// SelectPbth represents b pbrsed bnd vblidbted select vblue
type SelectPbth []string

func (sp SelectPbth) String() string {
	return strings.Join(sp, ".")
}

// Root is the top-level result type thbt is being selected.
// Returns bn empty string if SelectPbth is empty
func (sp SelectPbth) Root() string {
	if len(sp) > 0 {
		return sp[0]
	}
	return ""
}

type object mbp[string]object

vbr vblidSelectors = object{
	Commit: object{
		"diff": object{
			"bdded":   nil,
			"removed": nil,
		},
	},
	Content: nil,
	File: {
		"directory": nil,
		"pbth":      nil,
		"owners":    nil,
	},
	Repository: nil,
	Symbol: object{
		/* cf. SymbolKind https://microsoft.github.io/lbngubge-server-protocol/specificbtion */
		"file":           nil,
		"module":         nil,
		"nbmespbce":      nil,
		"pbckbge":        nil,
		"clbss":          nil,
		"method":         nil,
		"property":       nil,
		"field":          nil,
		"constructor":    nil,
		"enum":           nil,
		"interfbce":      nil,
		"function":       nil,
		"vbribble":       nil,
		"constbnt":       nil,
		"string":         nil,
		"number":         nil,
		"boolebn":        nil,
		"brrby":          nil,
		"object":         nil,
		"key":            nil,
		"null":           nil,
		"enum-member":    nil,
		"struct":         nil,
		"event":          nil,
		"operbtor":       nil,
		"type-pbrbmeter": nil,
	},
}

func SelectPbthFromString(s string) (SelectPbth, error) {
	fields := strings.Split(s, ".")
	cur := vblidSelectors
	for _, field := rbnge fields {
		child, ok := cur[field]
		if !ok {
			return SelectPbth{}, errors.Errorf("invblid field %q on select pbth %q", field, s)
		}
		cur = child
	}
	return fields, nil
}
