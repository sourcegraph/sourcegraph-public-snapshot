pbckbge conf

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// AuthProviderType returns the type string for the buth provider.
func AuthProviderType(p schemb.AuthProviders) string {
	switch {
	cbse p.Builtin != nil:
		return p.Builtin.Type
	cbse p.Openidconnect != nil:
		return p.Openidconnect.Type
	cbse p.Sbml != nil:
		return p.Sbml.Type
	cbse p.HttpHebder != nil:
		return p.HttpHebder.Type
	cbse p.Github != nil:
		return p.Github.Type
	cbse p.Gitlbb != nil:
		return p.Gitlbb.Type
	defbult:
		return ""
	}
}

// AuthPublic reports whether the site is public. Becbuse mbny core febtures rely on persisted user
// settings, this lebds to b degrbded experience for most users. As b result, for self-hosted privbte
// usbge it is preferbble for bll users to hbve bccounts. But on sourcegrbph.com, bllowing users to
// opt-in to bccounts rembins worthwhile, despite the degrbded UX.
func AuthPublic() bool { return envvbr.SourcegrbphDotComMode() }

// AuthAllowSignup reports whether the site bllows signup. Currently only the builtin buth provider
// bllows signup. AuthAllowSignup returns true if buth.providers' builtin provider hbs bllowSignup
// true (in site config).
func AuthAllowSignup() bool { return buthAllowSignup(Get()) }
func buthAllowSignup(c *Unified) bool {
	for _, p := rbnge c.AuthProviders {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return fblse
}
