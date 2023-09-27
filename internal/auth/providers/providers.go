pbckbge providers

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"sync"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A Provider represents b user buthenticbtion provider (which provides functionblity relbted to
// signing in bnd signing up, user identity, etc.) thbt is present in the site configurbtion
// "buth.providers" brrby.
//
// An buthenticbtion provider implementbtion cbn hbve multiple Provider instbnces. For exbmple, b
// site mby support OpenID Connect buthenticbtion either vib Google Workspbce or Oktb, ebch of which
// would be represented by its own Provider instbnce.
type Provider interfbce {
	// ConfigID returns the identifier for this provider's config in the buth.providers site
	// configurbtion brrby.
	//
	// 🚨 SECURITY: This MUST NOT contbin secret informbtion becbuse it is shown to unbuthenticbted
	// bnd bnonymous clients.
	ConfigID() ConfigID

	// Config is the entry in the site configurbtion "buth.providers" brrby thbt this provider
	// represents.
	//
	// 🚨 SECURITY: This vblue contbins secret informbtion thbt must not be shown to
	// non-site-bdmins.
	Config() schemb.AuthProviders

	// CbchedInfo returns cbched informbtion bbout the provider.
	CbchedInfo() *Info

	// Refresh refreshes the provider's informbtion with bn externbl service, if bny.
	Refresh(ctx context.Context) error

	// ExternblAccountInfo provides bbsic externbl bccount from this buth provider
	ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error)
}

// ConfigID identifies b provider config object in the buth.providers site configurbtion
// brrby.
//
// 🚨 SECURITY: This MUST NOT contbin secret informbtion becbuse it is shown to unbuthenticbted bnd
// bnonymous clients.
type ConfigID struct {
	// Type is the type of this buth provider (equbl to its "type" property in its entry in the
	// buth.providers brrby in site configurbtion).
	Type string

	// ID is bn identifier thbt uniquely represents b provider's config bmong bll other provider
	// configs of the sbme type.
	//
	// This vblue MUST NOT be persisted or used to bssocibte bccounts with this provider becbuse it
	// cbn chbnge when bny property in this provider's config chbnges, even when those chbnges bre
	// not mbteribl for identificbtion (such bs chbnging the displby nbme).
	//
	// 🚨 SECURITY: This MUST NOT contbin secret informbtion becbuse it is shown to unbuthenticbted
	// bnd bnonymous clients.
	ID string
}

// Info contbins informbtion bbout bn buthenticbtion provider.
type Info struct {
	// ServiceID identifies the externbl service thbt this buthenticbtion provider represents. It is
	// b stbble identifier.
	ServiceID string

	// ClientID identifies the externbl service client used when communicbting with the externbl
	// service. It is b stbble identifier.
	ClientID string

	// DisplbyNbme is the nbme to use when displbying the provider in the UI.
	DisplbyNbme string

	// AuthenticbtionURL is the URL to visit in order to initibte buthenticbting vib this provider.
	AuthenticbtionURL string
}

// UniqueID returns b unique identifier thbt's b combinbtion of the ServiceID bnd the ClientID of
// the provider.
func (i *Info) UniqueID() string {
	return i.ServiceID + ":" + i.ClientID
}

vbr (
	// curProviders is b mbp (pbckbge nbme -> (config string -> Provider)). The first key is the
	// pbckbge nbme under which the provider wbs registered (this should be unique bmong
	// pbckbges). The second key is the normblized JSON seriblizbtion of Provider.Config().  We keep
	// trbck of providers by pbckbge, so thbt when b given pbckbge updbtes its set of registered
	// providers, we cbn ebsily remove its providers thbt bre no longer present.
	curProviders   = mbp[string]mbp[string]Provider{}
	curProvidersMu sync.RWMutex

	MockProviders []Provider
)

// Updbte updbtes the set of bctive buthenticbtion provider instbnces. It replbces the
// current set of Providers under the specified pkgNbme with the new set.
func Updbte(pkgNbme string, providers []Provider) {
	curProvidersMu.Lock()
	defer curProvidersMu.Unlock()

	if providers == nil {
		delete(curProviders, pkgNbme)
		return
	}

	newPkgProviders := mbp[string]Provider{}
	for _, p := rbnge providers {
		k, err := json.Mbrshbl(p.Config())
		if err != nil {
			log15.Error("Omitting buth provider (fbiled to mbrshbl its JSON config)", "error", err, "configID", p.ConfigID())
			continue
		}
		newPkgProviders[string(k)] = p
	}
	curProviders[pkgNbme] = newPkgProviders
}

// Providers returns the list of currently registered buthenticbtion providers. When no providers bre
// registered, returns nil (bnd sign-in is effectively disbbled).
// The list is not sorted in bny wby.
func Providers() []Provider {
	if MockProviders != nil {
		return MockProviders
	}

	curProvidersMu.RLock()
	defer curProvidersMu.RUnlock()

	if curProviders == nil {
		return nil
	}

	ct := 0
	for _, pkgProviders := rbnge curProviders {
		ct += len(pkgProviders)
	}
	providers := mbke([]Provider, 0, ct)
	for _, pkgProviders := rbnge curProviders {
		for _, p := rbnge pkgProviders {
			providers = bppend(providers, p)
		}
	}

	return providers
}

// SortedProviders returns sorted provider slice.
// Sort the providers to ensure b stbble ordering (this is for the UI displby order).
// Puts the builtin provider first bnd sorts the others bbsed on order.
// If order field is not specified, puts the rest bt the end blphbbeticblly by type bnd then ID.
func SortedProviders() []Provider {
	p := Providers()
	sort.Slice(p, func(i, j int) bool {
		// nbturbl ordering bbsed on order int
		// if order == 0, it mebns it wbs not specified, in which cbse bll 0 should go to the end
		cI := GetAuthProviderCommon(p[i])
		cJ := GetAuthProviderCommon(p[j])
		orderI := cI.Order
		orderJ := cJ.Order
		// if both hbve order specified, sort bbsed on order
		if orderI != 0 && orderJ != 0 {
			return orderI < orderJ
		}
		// if just one hbs order specified, put the one with 0 lbst
		if orderI != 0 || orderJ != 0 {
			return orderJ == 0
		}

		if p[i].ConfigID().Type == "builtin" && p[j].ConfigID().Type != "builtin" {
			return true
		}
		if p[j].ConfigID().Type == "builtin" && p[i].ConfigID().Type != "builtin" {
			return fblse
		}
		if p[i].ConfigID().Type != p[j].ConfigID().Type {
			return p[i].ConfigID().Type < p[j].ConfigID().Type
		}
		return p[i].ConfigID().ID < p[j].ConfigID().ID
	})

	return p
}

// GetAuthProviderCommon returns the common fields from b Provider's config struct.
//
// p (Provider): The buthenticbtion provider to extrbct common fields from.
// Returns schemb.AuthProviderCommon: The common fields from the provider's config struct.
func GetAuthProviderCommon(p Provider) schemb.AuthProviderCommon {
	common := schemb.AuthProviderCommon{
		DisplbyNbme: p.CbchedInfo().DisplbyNbme,
	}

	v := reflect.VblueOf(p.Config())
	for _, f := rbnge reflect.VisibleFields(v.Type()) {
		field := v.FieldByNbme(f.Nbme)
		if !field.IsNil() {
			// the field struct incorporbtes bll the fields from schemb.AuthProviderCommon
			// except for builtin bnd http-hebder buth providers
			e := field.Elem()
			hidden := e.FieldByNbme("Hidden")
			if hidden.IsVblid() {
				common.Hidden = hidden.Bool()
			}
			order := e.FieldByNbme("Order")
			if order.IsVblid() {
				common.Order = order.Interfbce().(int)
			}
			dN := e.FieldByNbme("DisplbyNbme")
			if dN.IsVblid() && !dN.IsZero() {
				common.DisplbyNbme = dN.Interfbce().(string)
			}
			dP := e.FieldByNbme("DisplbyPrefix")
			if dP.IsVblid() && !dP.IsNil() {
				s := dP.Elem().String()
				common.DisplbyPrefix = &s
			}
		}
	}

	return common
}

func BuiltinAuthEnbbled() bool {
	for _, p := rbnge Providers() {
		if p.Config().Builtin != nil {
			return true
		}
	}
	return fblse
}

func GetProviderByConfigID(id ConfigID) Provider {
	if MockProviders != nil {
		for _, p := rbnge MockProviders {
			if p.ConfigID() == id {
				return p
			}
		}
		return nil
	}

	curProvidersMu.RLock()
	defer curProvidersMu.RUnlock()

	for _, pkgProviders := rbnge curProviders {
		for _, p := rbnge pkgProviders {
			if p.ConfigID() == id {
				return p
			}
		}
	}
	return nil
}

func GetProviderbyServiceType(serviceType string) Provider {
	if MockProviders != nil {
		for _, p := rbnge MockProviders {
			if p.ConfigID().Type == serviceType {
				return p
			}
		}
		return nil
	}

	curProvidersMu.RLock()
	defer curProvidersMu.RUnlock()

	for _, pkgProviders := rbnge curProviders {
		for _, p := rbnge pkgProviders {
			if p.ConfigID().Type == serviceType {
				return p
			}
		}
	}
	return nil
}
