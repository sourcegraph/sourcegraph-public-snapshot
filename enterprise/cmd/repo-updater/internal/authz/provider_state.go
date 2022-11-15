package authz

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
)

// providerState describes the state of a provider during a permissions sync.
type providerState struct {
	ProviderID   string
	ProviderType string

	// One of "ERROR" or "SUCCESS"
	State   string
	Message string
}

func newProviderState(provider authz.Provider, err error, action string) providerState {
	if err != nil {
		return providerState{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			State:        "ERROR",
			Message:      fmt.Sprintf("%s: %s", action, err.Error()),
		}
	} else {
		return providerState{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			State:        "SUCCESS",
			Message:      action,
		}
	}
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
