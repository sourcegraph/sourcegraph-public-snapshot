package validation

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	conf.ContributeValidator(codyProConfigValidator)
}

// validOrigin returns whether or not the supplied string is a valid HTTP origin,
// that is a scheme, domain, and optional port ONLY.
func validOrigin(origin string) error {
	// URL parsing will drop any "/" suffix. But we require it to
	// be missing from the string config value since we concatenate
	// assuming it will not be present.
	if strings.HasSuffix(origin, "/") {
		return errors.New("URL must not end in a tailing slash")
	}

	parsedURL, err := url.Parse(origin)
	if err != nil {
		return errors.Wrap(err, "invalid URL")
	}

	// NOTE: We do not require the scheme be https to support
	// connecting to local instances for development.
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("invalid scheme")
	}

	if parsedURL.RawPath != "" {
		return errors.New("URL path must be omitted")
	}
	if parsedURL.RawQuery != "" {
		return errors.New("URL query must be omitted")
	}
	if parsedURL.Fragment != "" {
		return errors.New("URL fragment must be omitted")
	}

	return nil
}

func codyProConfigValidator(q conftypes.SiteConfigQuerier) conf.Problems {
	// Noop unless the site config actually contains Cody Pro settings.
	dotcomConfig := q.SiteConfig().Dotcom
	if dotcomConfig == nil {
		return nil
	}
	codyProConfig := dotcomConfig.CodyProConfig
	if codyProConfig == nil {
		return nil
	}

	var problems []string
	// Confirm the Stripe Publishable Key has the expected prefix.
	if !strings.HasPrefix(codyProConfig.StripePublishableKey, "pk_test_") && !strings.HasPrefix(codyProConfig.StripePublishableKey, "pk_live_") {
		problems = append(problems, "codyProConfig.stripePublishableKey is invalid.")
	}

	if err := validOrigin(codyProConfig.SamsBackendOrigin); err != nil {
		problems = append(problems, "codyProConfig.samsBackendOrigin is invalid: "+err.Error())
	}
	if err := validOrigin(codyProConfig.SscBackendOrigin); err != nil {
		problems = append(problems, "codyProConfig.sscBackendOrigin is invalid: "+err.Error())
	}

	// Bundle and return.
	if len(problems) > 0 {
		return conf.NewSiteProblems(problems...)
	}
	return nil
}
