package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PermissionSyncCodeHostState describes the state of a provider during an authz sync job.
type PermissionSyncCodeHostState struct {
	ProviderID   string `json:"provider_id"`
	ProviderType string `json:"provider_type"`

	// Status is one of "ERROR" or "SUCCESS".
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e *PermissionSyncCodeHostState) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &e)
}

func (e PermissionSyncCodeHostState) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func NewProviderStatus(provider authz.Provider, err error, action string) PermissionSyncCodeHostState {
	if err != nil {
		return PermissionSyncCodeHostState{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			Status:       "ERROR",
			Message:      fmt.Sprintf("%s: %s", action, err.Error()),
		}
	} else {
		return PermissionSyncCodeHostState{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			Status:       "SUCCESS",
			Message:      action,
		}
	}
}

type CodeHostStatusesSet []PermissionSyncCodeHostState

// SummaryField generates a single log field that summarizes the state of all providers.
func (ps CodeHostStatusesSet) SummaryField() log.Field {
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
