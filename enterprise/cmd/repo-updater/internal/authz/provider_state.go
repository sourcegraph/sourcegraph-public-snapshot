package authz

import (
	"fmt"

	"github.com/sourcegraph/log"
)

// providerState describes the state of a provider during a permissions sync.
type providerState struct {
	ProviderID   string
	ProviderType string

	// One of "ERROR" or "SUCCESS"
	State   string
	Message string
}

type providerStatesSet []providerState

// SummaryField generates a single log field that summarizes the state of all providers.
func (ps providerStatesSet) SummaryField() log.Field {
	var (
		errored   []log.Field
		succeeded []log.Field
	)
	for _, p := range ps {
		key := fmt.Sprintf("%s:%s", p.ProviderType, p.ProviderID)
		// first add errored providers to fields
		switch p.State {
		case "ERROR":
			errored = append(errored, log.String(
				key,
				p.Message,
			))
		case "SUCCESS":
			succeeded = append(errored, log.String(
				key,
				p.Message,
			))
		}
	}
	return log.Object("providers",
		log.Object("state.error", errored...),
		log.Object("state.success", succeeded...))
}
