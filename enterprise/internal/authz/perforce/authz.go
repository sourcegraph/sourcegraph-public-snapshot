package perforce

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Perforce authz providers derived from
// the connections. It also returns any validation problems with the config,
// separating these into "serious problems" and "warnings". "Serious problems"
// are those that should make Sourcegraph set authz.allowAccessByDefault to
// false. "Warnings" are all other validation problems.
func NewAuthzProviders(conns []*types.PerforceConnection) (ps []authz.Provider, problems []string, warnings []string) {
	for _, c := range conns {
		p, err := newAuthzProvider(c.URN, c.Authorization, c.P4Port, c.P4User, c.P4Passwd)
		if err != nil {
			problems = append(problems, err.Error())
		} else if p != nil {
			ps = append(ps, p)
		}
	}

	for _, p := range ps {
		for _, problem := range p.Validate() {
			warnings = append(warnings, fmt.Sprintf("Perforce config for %s was invalid: %s", p.ServiceID(), problem))
		}
	}

	return ps, problems, warnings
}

func newAuthzProvider(
	urn string,
	a *schema.PerforceAuthorization,
	host, user, password string,
) (authz.Provider, error) {
	if a == nil {
		return nil, nil
	}

	return NewProvider(urn, host, user, password), nil
}

// ValidateAuthz validates the authorization fields of the given Perforce
// external service config.
func ValidateAuthz(cfg *schema.PerforceConnection) error {
	_, err := newAuthzProvider("", cfg.Authorization, cfg.P4Port, cfg.P4User, cfg.P4Passwd)
	return err
}
