package conf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/schema"
)

// Validator is a function that is called to validate the site configuration. Any strings that
// it returns are shown as validation problems.
type Validator = func(SiteConfiguration) (problems []string)

// ContributeValidator adds the site configuration validator function to the validation process. It
// is called to validate site configuration. Any strings it returns are shown as validation
// problems.
//
// It may only be called at init time.
func ContributeValidator(f Validator) {
	contributedValidators = append(contributedValidators, f)
}

var contributedValidators = []Validator{
	// Seeded with these validators by default.
	validateEmail,
	validatePhabricator,
	validateBitbucketServer,
	validateGitlab,
}

func validateCustomBasicRaw(normalizedInput []byte) (problems []string, err error) {
	var basic schema.BasicSiteConfiguration
	if err := json.Unmarshal(normalizedInput, &basic); err != nil {
		return nil, err
	}
	return validateCustom(SiteConfiguration{basic, schema.CoreSiteConfiguration{}}), nil
}

func validateCustomCoreRaw(normalizedInput []byte) (problems []string, err error) {
	var core schema.CoreSiteConfiguration
	if err := json.Unmarshal(normalizedInput, &core); err != nil {
		return nil, err
	}
	return validateCustom(SiteConfiguration{schema.BasicSiteConfiguration{}, core}), nil
}

// validateCustom validates the site config using custom validation steps that are not
// able to be expressed in the JSON Schema.
func validateCustom(cfg SiteConfiguration) (problems []string) {
	for _, f := range contributedValidators {
		problems = append(problems, f(cfg)...)
	}

	return problems
}

func validateEmail(c SiteConfiguration) (problems []string) {
	hasSMTP := c.EmailSmtp != nil
	hasSMTPAuth := c.EmailSmtp != nil && c.EmailSmtp.Authentication != "none"

	if hasSMTP && c.EmailAddress == "" {
		problems = append(problems, "should set email.address because email.smtp is set")
	}

	if hasSMTPAuth && (c.EmailSmtp.Username == "" && c.EmailSmtp.Password == "") {
		problems = append(problems, "must set email.smtp username and password for email.smtp authentication")
	}

	return problems
}

func validatePhabricator(c SiteConfiguration) (problems []string) {
	for _, p := range c.Phabricator {
		if len(p.Repos) == 0 && p.Token == "" {
			problems = append(problems, `each phabricator instance must have either "token" or "repos" set`)
		}
	}

	return problems
}

func validateBitbucketServer(c SiteConfiguration) (problems []string) {
	for _, bb := range c.BitbucketServer {
		if bb.Token != "" && bb.Password != "" {
			problems = append(problems, "for Bitbucket Server, specify either a token or a username/password, not both")
		} else if bb.Token == "" && bb.Username == "" && bb.Password == "" {
			problems = append(problems, "for Bitbucket Server, you must specify either a token or a username/password to authenticate")
		}
	}

	return problems
}

func validateGitlab(c SiteConfiguration) (problems []string) {
	for _, g := range c.Gitlab {
		if strings.Contains(g.Url, "example.com") {
			problems = append(problems, fmt.Sprintf(`invalid GitLab URL detected: %s (did you forget to remove "example.com"?)`, g.Url))
		}
	}

	return problems
}

// TestValidator is an exported helper function for other packages to test their contributed
// validators (registered with ContributeValidator). It should only be called by tests.
func TestValidator(t interface {
	Errorf(format string, args ...interface{})
	Helper()
}, c SiteConfiguration, f func(SiteConfiguration) []string, wantProblems []string) {
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
