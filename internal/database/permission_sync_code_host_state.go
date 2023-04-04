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
	ProviderID   string         `json:"provider_id"`
	ProviderType string         `json:"provider_type"`
	Status       CodeHostStatus `json:"status"`
	Message      string         `json:"message"`
}

type CodeHostStatus string

const (
	CodeHostStatusSuccess CodeHostStatus = "SUCCESS"
	CodeHostStatusError   CodeHostStatus = "ERROR"
)

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
			Status:       CodeHostStatusError,
			Message:      fmt.Sprintf("%s: %s", action, err.Error()),
		}
	} else {
		return PermissionSyncCodeHostState{
			ProviderID:   provider.ServiceID(),
			ProviderType: provider.ServiceType(),
			Status:       CodeHostStatusSuccess,
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
		case CodeHostStatusError:
			errored = append(errored, log.String(
				key,
				p.Message,
			))
		case CodeHostStatusSuccess:
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

// CountStatuses returns 3 integers: numbers of total, successful and failed
// statuses consisted in given CodeHostStatusesSet.
func (ps CodeHostStatusesSet) CountStatuses() (total, success, failed int) {
	total = len(ps)
	for _, state := range ps {
		if state.Status == CodeHostStatusSuccess {
			success++
		} else {
			failed++
		}
	}
	return
}
