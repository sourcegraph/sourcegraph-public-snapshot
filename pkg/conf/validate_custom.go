package conf

import (
	"encoding/json"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

// ContributeValidator adds the site configuration validator function to the validation process. It
// is called to validate site configuration. Any strings it returns are shown as validation
// problems.
//
// It may only be called at init time.
func ContributeValidator(f func(Unified) (problems []string)) {
	contributedValidators = append(contributedValidators, f)
}

var contributedValidators []func(Unified) []string

func validateCustomRaw(normalizedInput conftypes.RawUnified) (problems []string, err error) {
	var cfg Unified
	if err := json.Unmarshal([]byte(normalizedInput.Critical), &cfg.Critical); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(normalizedInput.Site), &cfg.SiteConfiguration); err != nil {
		return nil, err
	}
	return validateCustom(cfg), nil
}

// validateCustom validates the site config using custom validation steps that are not
// able to be expressed in the JSON Schema.
func validateCustom(cfg Unified) (problems []string) {

	invalid := func(msg string) {
		problems = append(problems, msg)
	}

	// Auth provider config validation is contributed by the
	// github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/... packages (using
	// ContributeValidator).

	{
		hasSMTP := cfg.EmailSmtp != nil
		hasSMTPAuth := cfg.EmailSmtp != nil && cfg.EmailSmtp.Authentication != "none"
		if hasSMTP && cfg.EmailAddress == "" {
			invalid(`should set email.address because email.smtp is set`)
		}
		if hasSMTPAuth && (cfg.EmailSmtp.Username == "" && cfg.EmailSmtp.Password == "") {
			invalid(`must set email.smtp username and password for email.smtp authentication`)
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
	Errorf(format string, args ...interface{})
	Helper()
}, c Unified, f func(Unified) []string, wantProblems []string) {
	t.Helper()
	problems := f(c)
	wantSet := make(map[string]struct{}, len(wantProblems))
	for _, p := range wantProblems {
		wantSet[p] = struct{}{}
	}
	for _, p := range problems {
		var found bool
		for ps := range wantSet {
			if strings.Contains(p, ps) {
				delete(wantSet, ps)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("got unexpected error %q", p)
		}
	}
	if len(wantSet) > 0 {
		t.Errorf("got no matches for expected error substrings %q", wantSet)
	}
}
