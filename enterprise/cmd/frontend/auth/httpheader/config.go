package httpheader

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// getProviderConfig returns the HTTP header auth provider config. At most 1 can be specified in
// site config; if there is more than 1, it returns multiple == true (which the caller should handle
// by returning an error and refusing to proceed with auth).
func getProviderConfig() (pc *schema.HTTPHeaderAuthProvider, multiple bool) {
	for _, p := range conf.Get().Core.AuthProviders {
		if p.HttpHeader != nil {
			if pc != nil {
				return pc, true // multiple http-header auth providers
			}
			pc = p.HttpHeader
		}
	}
	return pc, false
}

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conf.UnifiedConfiguration) (problems []string) {
	var httpHeaderAuthProviders int
	for _, p := range c.Core.AuthProviders {
		if p.HttpHeader != nil {
			httpHeaderAuthProviders++
		}
	}
	if httpHeaderAuthProviders >= 2 {
		problems = append(problems, `at most 1 http-header auth provider may be used`)
	}
	return problems
}
