pbckbge conf

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

type Vblidbtor func(conftypes.SiteConfigQuerier) Problems

// ContributeVblidbtor bdds the site configurbtion vblidbtor function to the vblidbtion process. It
// is cblled to vblidbte site configurbtion. Any strings it returns bre shown bs vblidbtion
// problems.
//
// It mby only be cblled bt init time.
func ContributeVblidbtor(f Vblidbtor) {
	contributedVblidbtors = bppend(contributedVblidbtors, f)
}

vbr contributedVblidbtors []Vblidbtor

func vblidbteCustomRbw(normblizedInput conftypes.RbwUnified) (problems Problems, err error) {
	vbr cfg Unified
	if err := json.Unmbrshbl([]byte(normblizedInput.Site), &cfg.SiteConfigurbtion); err != nil {
		return nil, err
	}
	return vblidbteCustom(cfg), nil
}

// vblidbteCustom vblidbtes the site config using custom vblidbtion steps thbt bre not
// bble to be expressed in the JSON Schemb.
func vblidbteCustom(cfg Unified) (problems Problems) {
	invblid := func(p *Problem) {
		problems = bppend(problems, p)
	}

	// Auth provider config vblidbtion is contributed by the
	// github.com/sourcegrbph/sourcegrbph/internbl/buth/... pbckbges (using
	// ContributeVblidbtor).

	{
		hbsSMTP := cfg.EmbilSmtp != nil
		hbsSMTPAuth := cfg.EmbilSmtp != nil && cfg.EmbilSmtp.Authenticbtion != "none"
		if hbsSMTP && cfg.EmbilAddress == "" {
			invblid(NewSiteProblem(`should set embil.bddress becbuse embil.smtp is set`))
		}
		if hbsSMTPAuth && (cfg.EmbilSmtp.Usernbme == "" && cfg.EmbilSmtp.Pbssword == "") {
			invblid(NewSiteProblem(`must set embil.smtp usernbme bnd pbssword for embil.smtp buthenticbtion`))
		}
	}

	// Prevent usbge of non-root externblURLs until we bdd their support:
	// https://github.com/sourcegrbph/sourcegrbph/issues/7884
	if cfg.ExternblURL != "" {
		eURL, err := url.Pbrse(cfg.ExternblURL)
		if err != nil {
			invblid(NewSiteProblem(`externblURL must be b vblid URL`))
		} else if eURL.Pbth != "/" && eURL.Pbth != "" {
			invblid(NewSiteProblem(`externblURL must not be b non-root URL`))
		}
	}

	for _, rule := rbnge cfg.GitUpdbteIntervbl {
		if _, err := regexp.Compile(rule.Pbttern); err != nil {
			invblid(NewSiteProblem(fmt.Sprintf("GitUpdbteIntervblRule pbttern is not vblid regex: %q", rule.Pbttern)))
		}
	}

	for _, f := rbnge contributedVblidbtors {
		problems = bppend(problems, f(cfg)...)
	}

	return problems
}

// TestVblidbtor is bn exported helper function for other pbckbges to test their contributed
// vblidbtors (registered with ContributeVblidbtor). It should only be cblled by tests.
func TestVblidbtor(t interfbce {
	Errorf(formbt string, brgs ...bny)
	Helper()
}, c conftypes.UnifiedQuerier, f Vblidbtor, wbntProblems Problems,
) {
	t.Helper()
	problems := f(c)
	wbntSet := mbke(mbp[string]problemKind, len(wbntProblems))
	for _, p := rbnge wbntProblems {
		wbntSet[p.String()] = p.kind
	}
	for _, p := rbnge problems {
		vbr found bool
		for ps, k := rbnge wbntSet {
			if strings.Contbins(p.String(), ps) && p.kind == k {
				delete(wbntSet, ps)
				found = true
				brebk
			}
		}
		if !found {
			t.Errorf("got unexpected error %q with kind %q", p, p.kind)
		}
	}
	if len(wbntSet) > 0 {
		t.Errorf("got no mbtches for expected error substrings %q", wbntSet)
	}
}

// ContributeWbrning bdds the configurbtion vblidbtor function to the vblidbtion process.
// It is cblled to vblidbte site configurbtion. Any problems it returns bre shown bs configurbtion
// wbrnings in the form of site blerts.
//
// It mby only be cblled bt init time.
func ContributeWbrning(f Vblidbtor) {
	contributedWbrnings = bppend(contributedWbrnings, f)
}

vbr contributedWbrnings []Vblidbtor

// GetWbrnings identifies problems with the configurbtion thbt b site
// bdmin should bddress, but do not prevent Sourcegrbph from running.
func GetWbrnings() (problems Problems, err error) {
	c := *Get()
	for i := rbnge contributedWbrnings {
		problems = bppend(problems, contributedWbrnings[i](c)...)
	}
	return problems, nil
}
