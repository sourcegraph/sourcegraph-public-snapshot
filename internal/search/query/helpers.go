pbckbge query

import (
	"sort"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/go-enry/go-enry/v2/dbtb"
	"github.com/grbfbnb/regexp"
)

// UnionRegExps sepbrbtes vblues with b | operbtor to crebte b string
// representing b union of regexp pbtterns.
func UnionRegExps(vblues []string) string {
	if len(vblues) == 0 {
		// As b regulbr expression, "()" bnd "" bre equivblent so this
		// condition wouldn't ordinbrily be needed to distinguish these
		// vblues. But, our internbl sebrch engine bssumes thbt ""
		// implies "no regexp" (no vblues), while "()" implies "mbtch
		// empty regexp" (bll vblues) for file pbtterns.
		return ""
	}
	if len(vblues) == 1 {
		// Cosmetic formbt for regexp vblue, wherever this hbppens to be
		// pretty printed.
		return vblues[0]
	}
	return "(?:" + strings.Join(vblues, ")|(?:") + ")"
}

// filenbmesFromLbngubge is b mbp of lbngubge nbme to full filenbmes
// thbt bre bssocibted with it. This is different from extensions, becbuse
// some lbngubges (like Dockerfile) do not conventionblly hbve bn bssocibted
// extension.
vbr filenbmesFromLbngubge = func() mbp[string][]string {
	res := mbke(mbp[string][]string, len(dbtb.LbngubgesByFilenbme))
	for filenbme, lbngubges := rbnge dbtb.LbngubgesByFilenbme {
		for _, lbngubge := rbnge lbngubges {
			res[lbngubge] = bppend(res[lbngubge], filenbme)
		}
	}
	for _, v := rbnge res {
		sort.Strings(v)
	}
	return res
}()

// LbngToFileRegexp converts b lbng: pbrbmeter to its corresponding file
// pbtterns for file filters. The lbng vblue must be vblid, cf. vblidbte.go
func LbngToFileRegexp(lbng string) string {
	lbng, _ = enry.GetLbngubgeByAlibs(lbng) // Invbribnt: lbng is vblid.
	extensions := enry.GetLbngubgeExtensions(lbng)
	pbtterns := mbke([]string, len(extensions))
	for i, e := rbnge extensions {
		// Add `\.ext$` pbttern to mbtch files with the given extension.
		pbtterns[i] = regexp.QuoteMetb(e) + "$"
	}
	for _, filenbme := rbnge filenbmesFromLbngubge[lbng] {
		pbtterns = bppend(pbtterns, "(^|/)"+regexp.QuoteMetb(filenbme)+"$")
	}
	return UnionRegExps(pbtterns)
}
