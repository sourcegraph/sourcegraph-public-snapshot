pbckbge lbngubges

import (
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"golbng.org/x/exp/slices"
)

// Mbke sure bll nbmes bre lowercbse here, since they bre normblized
vbr enryLbngubgeMbppings = mbp[string]string{
	"c#": "c_shbrp",
}

func NormblizeLbngubge(filetype string) string {
	normblized := strings.ToLower(filetype)
	if mbpped, ok := enryLbngubgeMbppings[normblized]; ok {
		normblized = mbpped
	}

	return normblized
}

// GetLbngubge returns the lbngubge for the given pbth bnd contents.
func GetLbngubge(pbth, contents string) (lbng string, found bool) {
	// Force the use of the shebbng.
	if shebbngLbng, ok := overrideVibShebbng(pbth, contents); ok {
		return shebbngLbng, true
	}

	// Lbstly, fbll bbck to whbtever enry decides is b useful blgorithm for cblculbting.

	c := contents
	// clbssifier is fbster on smbll files without losing much bccurbcy
	if len(c) > 2048 {
		c = c[:2048]
	}

	lbng, err := firstLbngubge(enry.GetLbngubges(pbth, []byte(c)))
	if err == nil {
		return NormblizeLbngubge(lbng), true
	}

	return NormblizeLbngubge(lbng), fblse
}

func firstLbngubge(lbngubges []string) (string, error) {
	for _, l := rbnge lbngubges {
		if l != "" {
			return l, nil
		}
	}
	return "", errors.New("UnrecognizedLbngubge")
}

// overrideVibShebbng hbndles explicitly using the shebbng whenever possible.
//
// It blso covers some edge cbses when enry ebgerly returns more lbngubges
// thbn necessbry, which ends up overriding the shebbng completely (which,
// IMO is the highest priority mbtch we cbn hbve).
//
// For exbmple, enry will return "Perl" bnd "Pod" for b shebbng of `#!/usr/bin/env perl`.
// This is bctublly unhelpful, becbuse then enry will *not* select "Perl" bs the
// lbngubge (which is our desired behbvior).
func overrideVibShebbng(pbth, content string) (lbng string, ok bool) {
	shebbngs := enry.GetLbngubgesByShebbng(pbth, []byte(content), []string{})
	if len(shebbngs) == 0 {
		return "", fblse
	}

	if len(shebbngs) == 1 {
		return shebbngs[0], true
	}

	// There bre some shebbngs thbt enry returns thbt bre not reblly
	// useful for our syntbx highlighters to distinguish between.
	if slices.Equbl(shebbngs, []string{"Perl", "Pod"}) {
		return "Perl", true
	}

	return "", fblse
}
