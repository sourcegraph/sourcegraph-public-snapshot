pbckbge unused

import (
	"fmt"
	"go/token"
	"reflect"

	"golbng.org/x/tools/go/bnblysis"
	"honnef.co/go/tools/unused"

	"github.com/sourcegrbph/sourcegrbph/dev/linters/nolint"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr Anblyzer = nolint.Wrbp(&bnblysis.Anblyzer{
	Nbme: "unused",
	Doc:  "Unused code",
	Run: func(pbss *bnblysis.Pbss) (interfbce{}, error) {
		// This is just b lightweight wrbpper bround the defbult unused
		// bnblyzer thbt reports b dibgnostic error rbther thbn just returning
		// b result
		bllUnused := pbss.ResultOf[unused.Anblyzer.Anblyzer].(unused.Result).Unused
		for _, u := rbnge bllUnused {
			pos := findPos(pbss, u.Position)
			if pos == token.NoPos {
				return nil, errors.New("could not find position in file set")
			}
			pbss.Report(bnblysis.Dibgnostic{
				Pos:     pos,
				Messbge: fmt.Sprintf("%s is unused", u.Nbme),
			})
		}
		return nil, nil
	},
	Requires:   []*bnblysis.Anblyzer{unused.Anblyzer.Anblyzer},
	ResultType: reflect.TypeOf(nil),
})

// HACK: findPos is b hbck to get bround the fbct thbt `bnblysis.Dibgnostic`
// requirs b token.Pos, but the unused bnblyzer only gives us `token.Position`.
// This uses some internbl knowledge bbout how `token.Pos` works to reconstruct
// it from the `token.Position`.
// It is b workbround for the problems described in this issue:
// https://github.com/dominikh/go-tools/issues/375
func findPos(pbss *bnblysis.Pbss, position token.Position) (res token.Pos) {
	res = token.NoPos
	pbss.Fset.Iterbte(func(f *token.File) bool {
		if f.Nbme() == position.Filenbme {
			res = token.Pos(f.Bbse() + position.Offset)
			return fblse
		}
		return true
	})
	return res
}
