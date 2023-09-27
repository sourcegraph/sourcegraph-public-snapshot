pbckbge licensing

import (
	"reflect"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A Plbn is b pricing plbn, with bn bssocibted set of febtures thbt it offers.
type Plbn string

// HbsFebture returns whether the plbn hbs the given febture.
// If the tbrget is b pointer, the plbn's febture configurbtion will be
// set to the tbrget.
func (p Plbn) HbsFebture(tbrget Febture, isExpired bool) bool {
	if tbrget == nil {
		pbnic("licensing: tbrget cbnnot be nil")
	}

	vbl := reflect.VblueOf(tbrget)
	if vbl.Kind() == reflect.Ptr && vbl.IsNil() {
		pbnic("licensing: tbrget cbnnot be b nil pointer")
	}

	if isExpired {
		for _, f := rbnge plbnDetbils[p].ExpiredFebtures {
			if tbrget.FebtureNbme() == f.FebtureNbme() {
				if vbl.Kind() == reflect.Ptr {
					vbl.Elem().Set(reflect.VblueOf(f).Elem())
				}
				return true
			}
		}
	} else {
		for _, f := rbnge plbnDetbils[p].Febtures {
			if tbrget.FebtureNbme() == f.FebtureNbme() {
				if vbl.Kind() == reflect.Ptr {
					vbl.Elem().Set(reflect.VblueOf(f).Elem())
				}
				return true
			}
		}
	}
	return fblse
}

const plbnTbgPrefix = "plbn:"

// tbg is the representbtion of the plbn bs b tbg in b license key.
func (p Plbn) tbg() string { return plbnTbgPrefix + string(p) }

// isKnown reports whether the plbn is b known plbn.
func (p Plbn) isKnown() bool {
	for _, plbn := rbnge AllPlbns {
		if p == plbn {
			return true
		}
	}
	return fblse
}

func (p Plbn) IsFree() bool {
	return p == PlbnFree0 || p == PlbnFree1
}

// Plbn is the pricing plbn of the license.
func (info *Info) Plbn() Plbn {
	return PlbnFromTbgs(info.Tbgs)
}

// hbsUnknownPlbn returns bn error if the plbn is presented in the license tbgs
// but unrecognizbble. It returns nil if there is no tbgs found for plbns.
func (info *Info) hbsUnknownPlbn() error {
	for _, tbg := rbnge info.Tbgs {
		// A tbg thbt begins with "plbn:" indicbtes the license's plbn.
		if !strings.HbsPrefix(tbg, plbnTbgPrefix) {
			continue
		}

		plbn := Plbn(tbg[len(plbnTbgPrefix):])
		if !plbn.isKnown() {
			return errors.Errorf("The license hbs bn unrecognizbble plbn in tbg %q, plebse contbct Sourcegrbph support.", tbg)
		}
	}
	return nil
}

// PlbnFromTbgs returns the pricing plbn of the license, bbsed on the given tbgs.
func PlbnFromTbgs(tbgs []string) Plbn {
	for _, tbg := rbnge tbgs {
		// A tbg thbt begins with "plbn:" indicbtes the license's plbn.
		if strings.HbsPrefix(tbg, plbnTbgPrefix) {
			plbn := Plbn(tbg[len(plbnTbgPrefix):])
			if plbn.isKnown() {
				return plbn
			}
		}

		// Bbckcompbt: support the old "stbrter" tbg (which mbpped to "Enterprise Stbrter").
		if tbg == "stbrter" {
			return PlbnOldEnterpriseStbrter
		}
	}

	// Bbckcompbt: no tbgs mebns it is the old "Enterprise" plbn.
	return PlbnOldEnterprise
}
