package v1

import (
	"net/url"
	"strings"

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
	// Basic validation because url.Parse is VERY generous
	if domain == "" {
		return "", errors.New("domain is empty")
	}
	if !strings.Contains(domain, ".") {
		return "", errors.New("domain does contain a '.'")
	}

	u, err := url.Parse(domain)
	if err != nil {
		return "", errors.Wrap(err, "domain is not a valid URL")
	}

	// If the parsing didn't find a scheme, assume HTTP and try again.
	if u.Scheme == "" && u.Host == "" {
		u, err = url.Parse("http://" + domain)
		if err != nil {
			return "", errors.Wrap(err, "invalid URL")
		}
	}

	return u.Host, nil
}
