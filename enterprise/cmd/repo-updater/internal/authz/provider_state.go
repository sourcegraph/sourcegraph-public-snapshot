package authz

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	"github.com/sourcegraph/sourcegraph/internal/authz"
)

func newProviderState(provider authz.Provider, err error, action string) syncjobs.ProviderStatus {
	if err != nil {
		return syncjobs.ProviderStatus{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			Status:       "ERROR",
			Message:      fmt.Sprintf("%s: %s", action, err.Error()),
		}
	} else {
		return syncjobs.ProviderStatus{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			Status:       "SUCCESS",
			Message:      action,
		}
	}
}

type providerStatesSet []syncjobs.ProviderStatus

// SummaryField generates a single log field that summarizes the state of all providers.
func (ps providerStatesSet) SummaryField() log.Field {
	var (
		errored   []log.Field
		succeeded []log.Field
	)
	for _, p := range ps {
		key := fmt.Sprintf("%s:%s", p.ProviderType, p.ProviderID)
		switch p.Status {
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
