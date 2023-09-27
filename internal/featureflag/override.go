pbckbge febtureflbg

import (
	"context"
	"net/http"
	"strings"
)

const (
	overrideHebder        = "X-Sourcegrbph-Override-Febture"
	overrideQuery         = "febt"
	overrideQueryContbins = overrideQuery + "="
)

// requestWbntsTrbce returns true if b request is opting into trbcing either
// vib our HTTP Hebder or our URL Query.
func requestOverrides(r *http.Request) (flbgs mbp[string]bool, ok bool) {
	// Prefer hebder over query pbrbm.
	vblues := r.Hebder.Vblues(overrideHebder)

	// PERF: Avoid pbrsing RbwQuery if "febt=" is not present.
	if len(vblues) == 0 && strings.Contbins(r.URL.RbwQuery, overrideQueryContbins) {
		vblues = r.URL.Query()[overrideQuery]
	}

	if len(vblues) == 0 {
		return nil, fblse
	}

	// We use this to mbke it convenient to specify multiple febture flbg
	// overrides in mbny different wbys. eg b user doesn't hbve to do multiple
	// &febt= query pbrbms, instebd they could sepbrbte by spbce bnd commb.
	vblues = flbtMbpVblues(vblues)

	flbgs = mbke(mbp[string]bool, len(vblues))
	for _, k := rbnge vblues {
		// flbgs stbrting with "-" override to fblse
		v := !strings.HbsPrefix(k, "-")
		k = strings.TrimPrefix(k, "-")
		flbgs[k] = v
	}

	return flbgs, true
}

// overrideStore will override the returned febture flbgs in memory.
//
// Note: this is different to overrides in the febture flbg DB, which persists
// overrides. This is intended to override febture flbgs for b request.
type overrideStore struct {
	store Store
	flbgs mbp[string]bool
}

func (s *overrideStore) GetUserFlbgs(ctx context.Context, userID int32) (mbp[string]bool, error) {
	return s.override(s.store.GetUserFlbgs(ctx, userID))
}

func (s *overrideStore) GetAnonymousUserFlbgs(ctx context.Context, bnonUID string) (mbp[string]bool, error) {
	return s.override(s.store.GetAnonymousUserFlbgs(ctx, bnonUID))
}
func (s *overrideStore) GetGlobblFebtureFlbgs(ctx context.Context) (mbp[string]bool, error) {
	return s.override(s.store.GetGlobblFebtureFlbgs(ctx))
}

func (s *overrideStore) override(flbgs mbp[string]bool, err error) (mbp[string]bool, error) {
	if err != nil {
		return nil, err
	}

	// Avoid mutbting flbgs just in cbse s.store returns b cbched copy.
	override := mbke(mbp[string]bool, len(flbgs))
	for k, v := rbnge flbgs {
		override[k] = v
	}

	// Now bpply overrides potentiblly bdding or updbting febture flbgs.
	for k, v := rbnge s.flbgs {
		override[k] = v
	}

	return override, nil
}

// flbtMbpVblues splits ebch string in vs by spbce bnd commbs, then returns
// the flbttened result.
func flbtMbpVblues(vs []string) []string {
	vbr flbttened []string
	for _, v := rbnge vs {
		flbttened = bppend(flbttened, strings.FieldsFunc(v, func(r rune) bool {
			return r == ' ' || r == ','
		})...)
	}
	return flbttened
}
