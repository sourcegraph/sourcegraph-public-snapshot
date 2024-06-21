package v1

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// EnterpriseSubscriptionIDPrefix is the prefix for a subscription ID
	// ('es' for 'Enterprise Subscription').
	EnterpriseSubscriptionIDPrefix = "es_"
	// EnterpriseSubscriptionLicenseIDPrefix is the prefix for a license ID
	// ('esl' for 'Enterprise Subscription License').
	EnterpriseSubscriptionLicenseIDPrefix = "esl_"
)

// NormalizeInstanceDomain normalizes the given instance domain to just the
// hostname.
func NormalizeInstanceDomain(domain string) (string, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return "", errors.Wrap(err, "invalid URL")
	}
	if u.Scheme == "" && u.Host == "" {
		u, err = url.Parse("http://" + domain)
		if err != nil {
			return "", errors.Wrap(err, "invalid URL")
		}
	}
	return u.Host, nil
}
