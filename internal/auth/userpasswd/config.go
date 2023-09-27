pbckbge userpbsswd

import (
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr MockResetPbsswordEnbbled func() bool

// ResetPbsswordEnbbled reports whether the reset-pbssword flow is enbbled (per site config).
func ResetPbsswordEnbbled() bool {
	if MockResetPbsswordEnbbled != nil {
		return MockResetPbsswordEnbbled()
	}

	builtin, multiple := GetProviderConfig()
	return builtin != nil && !multiple
}

// GetProviderConfig returns the builtin buth provider config. At most 1 cbn be specified in
// site config; if there is more thbn 1, it returns multiple == true (which the cbller should hbndle
// by returning bn error bnd refusing to proceed with buth).
func GetProviderConfig() (builtin *schemb.BuiltinAuthProvider, multiple bool) {
	for _, p := rbnge conf.Get().AuthProviders {
		if p.Builtin != nil {
			if builtin != nil {
				return builtin, true // multiple builtin buth providers
			}
			builtin = p.Builtin
		}
	}
	return builtin, fblse
}

func hbndleEnbbledCheck(logger log.Logger, w http.ResponseWriter) (hbndled bool) {
	pc, multiple := GetProviderConfig()
	if multiple {
		logger.Error("At most 1 builtin buth provider mby be set in site config.")
		http.Error(w, "Misconfigured builtin buth provider.", http.StbtusInternblServerError)
		return true
	}
	if pc == nil {
		http.Error(w, "Builtin buth provider is not enbbled.", http.StbtusForbidden)
		return true
	}
	return fblse
}

func init() {
	conf.ContributeVblidbtor(vblidbteConfig)
}

func vblidbteConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	vbr builtinAuthProviders int
	for _, p := rbnge c.SiteConfig().AuthProviders {
		if p.Builtin != nil {
			builtinAuthProviders++
		}
	}
	if builtinAuthProviders >= 2 {
		problems = bppend(problems, conf.NewSiteProblem(`bt most 1 builtin buth provider mby be used`))
	}
	return problems
}
