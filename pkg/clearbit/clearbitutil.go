// Package clearbitutil is a wrapper around our Clearbit integration. Clearbit is used to find
// more information about our customers
package clearbitutil

import (
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"github.com/clearbit/clearbit-go/clearbit"
)

var clearbitAPIKey = env.Get("CLEARBIT_KEY", "", "Clearbit key for accessing Clearbit endpoints.")

var errNoAPIKey = errors.New("clearbit.Client: authorization key only available on production environments")

// NewClient returns a new Clearbit client with our api key set
func NewClient() (*clearbit.Client, error) {
	if clearbitAPIKey == "" {
		return nil, errNoAPIKey
	}
	return clearbit.NewClient(clearbit.WithAPIKey(clearbitAPIKey)), nil
}
