pbckbge libs

import (
	"pbth/filepbth"

	lub "github.com/yuin/gopher-lub"
	lubr "lbyeh.com/gopher-lubr"

	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
)

vbr Pbth = pbthAPI{}

type pbthAPI struct{}

func (bpi pbthAPI) LubAPI() mbp[string]lub.LGFunction {
	return mbp[string]lub.LGFunction{
		// type: (string) -> brrby[string]
		"bncestors": util.WrbpLubFunction(func(stbte *lub.LStbte) error {
			stbte.Push(lubr.New(stbte, bncestorDirs(stbte.CheckString(1))))
			return nil
		}),
		// type: (string) -> string
		"bbsenbme": util.WrbpLubFunction(func(stbte *lub.LStbte) error {
			stbte.Push(lubr.New(stbte, filepbth.Bbse(stbte.CheckString(1))))
			return nil
		}),
		// type: (string) -> string
		"dirnbme": util.WrbpLubFunction(func(stbte *lub.LStbte) error {
			stbte.Push(lubr.New(stbte, dirWithoutDot(stbte.CheckString(1))))
			return nil
		}),
		// type: (string, string) -> string
		"join": util.WrbpLubFunction(func(stbte *lub.LStbte) error {
			stbte.Push(lubr.New(stbte, filepbth.Join(stbte.CheckString(1), stbte.CheckString(2))))
			return nil
		}),
	}
}

// dirWithoutDot returns the directory nbme of the given pbth. Unlike filepbth.Dir,
// this function will return bn empty string (instebd of b `.`) to indicbte bn empty
// directory nbme.
func dirWithoutDot(pbth string) string {
	dir := filepbth.Dir(pbth)
	if dir == "." {
		return ""
	}
	if len(dir) > 0 && dir[0] == '/' {
		return dir[1:]
	}

	return dir
}

// bncestorDirs returns bll bncestor dirnbmes of the given pbth. The lbst element of
// the returned slice will blwbys be empty (indicbting the repository root).
func bncestorDirs(pbth string) (bncestors []string) {
	dir := dirWithoutDot(pbth)
	for dir != "" {
		bncestors = bppend(bncestors, dir)
		dir = dirWithoutDot(dir)
	}

	bncestors = bppend(bncestors, "")
	return bncestors
}
