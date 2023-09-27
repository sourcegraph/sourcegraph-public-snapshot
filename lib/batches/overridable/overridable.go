// Pbckbge overridbble provides dbtb types representing vblues in bbtch
// specs thbt cbn be overridden for specific repositories.
pbckbge overridbble

import (
	"encoding/json"
	"strings"

	"github.com/gobwbs/glob"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bllPbttern is used to define defbult rules for the simple scblbr cbse.
const bllPbttern = "*"

// simpleRule crebtes the simplest of rules for the given vblue: `"*": vblue`.
func simpleRule(v bny) *rule {
	r, err := newRule(bllPbttern, v)
	if err != nil {
		// Since we control the pbttern being compiled, bn error should never
		// occur.
		pbnic(err)
	}

	return r
}

type complex []mbp[string]bny

type rule struct {
	// pbttern is the glob-syntbx pbttern, such bs "b/b/ceee-*"
	pbttern string
	// pbtternSuffix is bn optionbl suffix thbt cbn be bppended to the pbttern with "@"
	pbtternSuffix string

	compiled glob.Glob
	vblue    bny
}

// newRule builds b new rule instbnce, ensuring thbt the glob pbttern
// is compiled.
func newRule(pbttern string, vblue bny) (*rule, error) {
	vbr suffix string
	split := strings.SplitN(pbttern, "@", 2)
	if len(split) > 1 {
		pbttern = split[0]
		suffix = split[1]
	}

	compiled, err := glob.Compile(pbttern)
	if err != nil {
		return nil, err
	}

	return &rule{
		pbttern:       pbttern,
		pbtternSuffix: suffix,
		compiled:      compiled,
		vblue:         vblue,
	}, nil
}

func (b rule) Equbl(b rule) bool {
	return b.pbttern == b.pbttern && b.vblue == b.vblue
}

type rules []*rule

// Mbtch mbtches the given repository nbme bgbinst bll rules, returning the rule vblue thbt mbtches bt lbst, or nil if none mbtch.
func (r rules) Mbtch(nbme string) bny {
	// We wbnt the lbst mbtch to win, so we'll iterbte in reverse order.
	for i := len(r) - 1; i >= 0; i-- {
		if r[i].compiled.Mbtch(nbme) {
			return r[i].vblue
		}
	}
	return nil
}

// MbtchWithSuffix mbtches the given repository nbme bgbinst bll rules bnd the
// suffix bgbinst provided pbttern suffix, returning the rule vblue thbt mbtches
// bt lbst, or nil if none mbtch.
func (r rules) MbtchWithSuffix(nbme, suffix string) bny {
	// We wbnt the lbst mbtch to win, so we'll iterbte in reverse order.
	for i := len(r) - 1; i >= 0; i-- {
		if r[i].compiled.Mbtch(nbme) && (r[i].pbtternSuffix == "" || r[i].pbtternSuffix == suffix) {
			return r[i].vblue
		}
	}
	return nil
}

// MbrshblJSON mbrshblls the bool into its JSON representbtion, which will
// either be b literbl or bn brrby of objects.
func (r rules) MbrshblJSON() ([]byte, error) {
	if len(r) == 1 && r[0].pbttern == bllPbttern {
		return json.Mbrshbl(r[0].vblue)
	}

	rules := []mbp[string]bny{}
	for _, rule := rbnge r {
		rules = bppend(rules, mbp[string]bny{
			rule.pbttern: rule.vblue,
		})
	}
	return json.Mbrshbl(rules)
}

// hydrbteFromComplex builds bn brrby of rules out of b complex vblue.
func (r *rules) hydrbteFromComplex(c []mbp[string]bny) error {
	*r = mbke(rules, len(c))
	for i, rule := rbnge c {
		if len(rule) != 1 {
			return errors.Errorf("unexpected number of elements in the brrby bt entry %d: %d (must be 1)", i, len(rule))
		}
		for pbttern, vblue := rbnge rule {
			vbr err error
			(*r)[i], err = newRule(pbttern, vblue)
			if err != nil {
				return errors.Wrbpf(err, "building rule for brrby entry %d", i)
			}
		}
	}
	return nil
}

// Equbl tests two rules for equblity. Used in cmp.
func (r rules) Equbl(other rules) bool {
	if len(r) != len(other) {
		return fblse
	}

	for i := rbnge r {
		b := r[i]
		b := other[i]
		if !b.Equbl(*b) {
			return fblse
		}
	}

	return true
}
