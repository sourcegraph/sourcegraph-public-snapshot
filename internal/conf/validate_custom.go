package conf

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

type Validator func(conftypes.SiteConfigQuerier) Problems

// ContributeValidator adds the site configuration validator function to the validation process. It
// is called to validate site configuration. Any strings it returns are shown as validation
// problems.
//
// It may only be called at init time.
func ContributeValidator(f Validator) {
	contributedValidators = append(contributedValidators, f)
}

var contributedValidators []Validator

func validateCustomRaw(normalizedInput conftypes.RawUnified) (problems Problems, err error) {
	var cfg Unified
	if err := json.Unmarshal([]byte(normalizedInput.Site), &cfg.SiteConfiguration); err != nil {
		return nil, err
	}
	return validateCustom(cfg), nil
}

// validateCustom validates the site config using custom validation steps that are not
// able to be expressed in the JSON Schema.
func validateCustom(cfg Unified) (problems Problems) {
	invalid := func(p *Problem) {
		problems = append(problems, p)
	}

	// Auth provider config validation is contributed by the
	// github.com/sourcegraph/sourcegraph/internal/auth/... packages (using
	// ContributeValidator).

	{
		hasSMTP := cfg.EmailSmtp != nil
		hasSMTPAuth := cfg.EmailSmtp != nil && cfg.EmailSmtp.Authentication != "none"
		if hasSMTP && cfg.EmailAddress == "" {
			invalid(NewSiteProblem(`should set email.address because email.smtp is set`))
		}
		if hasSMTPAuth && (cfg.EmailSmtp.Username == "" && cfg.EmailSmtp.Password == "") {
			invalid(NewSiteProblem(`must set email.smtp username and password for email.smtp authentication`))
		}
	}

	// Prevent usage of non-root externalURLs until we add their support:
	// https://github.com/sourcegraph/sourcegraph/issues/7884
	if cfg.ExternalURL != "" {
		eURL, err := url.Parse(cfg.ExternalURL)
		if err != nil {
			invalid(NewSiteProblem(`externalURL must be a valid URL`))
		} else if eURL.Path != "/" && eURL.Path != "" {
			invalid(NewSiteProblem(`externalURL must not be a non-root URL`))
		}
	}

	for _, rule := range cfg.GitUpdateInterval {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			invalid(NewSiteProblem(fmt.Sprintf("GitUpdateIntervalRule pattern is not valid regex: %q", rule.Pattern)))
		}
	}

	for _, f := range contributedValidators {
		problems = append(problems, f(cfg)...)
	}

	return problems
}

// TestValidator is an exported helper function for other packages to test their contributed
// validators (registered with ContributeValidator). It should only be called by tests.
func TestValidator(t interface {
	Errorf(format string, args ...any)
	Helper()
}, c conftypes.UnifiedQuerier, f Validator, wantProblems Problems,
) {
	t.Helper()
	problems := f(c)
	wantSet := make(map[string]problemKind, len(wantProblems))
	for _, p := range wantProblems {
		wantSet[p.String()] = p.kind
	}
	for _, p := range problems {
		var found bool
		for ps, k := range wantSet {
			if strings.Contains(p.String(), ps) && p.kind == k {
				delete(wantSet, ps)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("got unexpected error %q with kind %q", p, p.kind)
		}
	}
	if len(wantSet) > 0 {
		t.Errorf("got no matches for expected error substrings %q", wantSet)
	}
}
