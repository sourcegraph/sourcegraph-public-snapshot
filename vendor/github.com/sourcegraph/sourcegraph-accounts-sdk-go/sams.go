package sams

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// envGetter is a helper interface for getting environment variables based on
// the MSP runtime environment, matching the type
// github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract.Env.
type envGetter interface {
	// Get returns the value with the given name. If no value was supplied in the
	// environment, the given default is used in its place. If no value is available,
	// an error is added to the validation errors list.
	Get(name, defaultValue, description string) string
	// GetOptional returns the value with the given name, or nil if no value is
	// available.
	GetOptional(name, description string) *string
}

// ConnConfig is the basic configuration for connecting to a Sourcegraph Accounts
// instance. Callers SHOULD use sams.NewConnConfigFromEnv(...) to construct this
// where possible to ensure connection configuration is unified across SAMS clients
// for ease of operation by Core Services team.
type ConnConfig struct {
	// ExternalURL is the configured default external ExternalURL of the relevant Sourcegraph
	// Accounts instance.
	ExternalURL string
	// APIURL is the URL to use for Sourcegraph Accounts API interactions. This
	// can be set to some internal URLs for private networking. If this is nil,
	// the client will fall back to ExternalURL instead.
	APIURL *string
}

const DefaultExternalURL = "https://accounts.sourcegraph.com"

// NewConnConfigFromEnv initializes configuration for connecting to Sourcegraph
// Accounts using default standards for loading environment variables. This
// allows the Core Services team to more easily configure access.
func NewConnConfigFromEnv(env envGetter) ConnConfig {
	return ConnConfig{
		ExternalURL: env.Get("SAMS_URL", DefaultExternalURL, "External URL of the connected SAMS instance"),
		APIURL:      env.GetOptional("SAMS_API_URL", "URL to use for connecting to the API of a SAMS instance instead of SAMS_URL"),
	}
}

func (c ConnConfig) Validate() error {
	if c.ExternalURL == "" {
		return errors.New("empty external URL")
	}
	if c.getAPIURL() == "" {
		return errors.New("evaluated API URL is empty")
	}
	return nil
}

func (c ConnConfig) getAPIURL() string {
	if c.APIURL != nil {
		return *c.APIURL
	}
	return c.ExternalURL
}
