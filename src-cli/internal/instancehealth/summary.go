package instancehealth

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
)

// Indicators are values from the Sourcegraph GraphQL API that help indicate the health of
// and instance.
type Indicators struct {
	Site struct {
		Configuration struct {
			ValidationMessages []string
		}
		Alerts []struct {
			Type    string
			Message string
		}
		MonitoringStatistics struct {
			Alerts []struct {
				Name    string
				Average float64
			}
		}
	}
	ExternalServices struct {
		Nodes []struct {
			Kind          string
			ID            string
			LastSyncError *string
			SyncJobs      struct {
				Nodes []struct {
					State      string
					FinishedAt time.Time
				}
			}
		}
	}
	PermissionsSyncJobs struct {
		Nodes []permissionSyncJob
	}
}

type permissionSyncJob struct {
	State          string
	FailureMessage string
	FinishedAt     time.Time
	CodeHostStates []permissionsProviderStatus
}

type permissionsProviderStatus struct {
	ProviderType string
	ProviderID   string
	Status       string
	Message      string
}

// GetIndicators retrieves summary data from a Sourcegraph instance's GraphQL API for
// assessing instance health.
func GetIndicators(ctx context.Context, client api.Client) (*Indicators, error) {
	var instanceHealth Indicators
	ok, err := client.NewQuery(`
		query InstanceHealthSummary {
			site {
				configuration {
					validationMessages
				}
				alerts {
					type
					message
				}
				monitoringStatistics(days:1) {
					alerts {
						name
						average
					}
				}
			}

			externalServices {
				nodes {
					kind
					id
					lastSyncError
					syncJobs(first:100) {
						nodes {
							state
							finishedAt
						}
					}
				}
			}

			permissionsSyncJobs(first:500) {
				nodes {
					state
					finishedAt
					failureMessage
					codeHostStates {
						providerType
						providerID
						status
						message
					}
				}
			}
		}
	`).Do(ctx, &instanceHealth)
	if err != nil {
		return nil, errors.Wrap(err, "get health data")
	} else if !ok {
		return nil, errors.New("received no data")
	}
	return &instanceHealth, nil
}
