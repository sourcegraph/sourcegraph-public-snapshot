package enforcement

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

// ExternalServicesStore is implemented by any type that can act as a
// repository for external services (e.g. GitHub, GitLab).
type ExternalServicesStore interface {
	Count(context.Context, database.ExternalServicesListOptions) (int, error)
}

// NewBeforeCreateExternalServiceHook enforces any per-tier validations prior to
// creating a new external service.
func NewBeforeCreateExternalServiceHook() func(ctx context.Context, store database.ExternalServiceStore) error {
	if !licensing.EnforceTiers {
		return nil
	}

	return func(ctx context.Context, store database.ExternalServiceStore) error {
		// Licenses are associated with features and resource limits according to
		// the current plan. We first need to determine the instance license, and then
		// extract the maximum external service count from it.
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return err
		}
		var maxExtSvcCount int
		if info != nil {
			maxExtSvcCount = info.Plan().MaxExternalServiceCount()
		} else {
			maxExtSvcCount = licensing.NoLicenseMaximumExternalServiceCount
		}

		// Next we'll grab the current count of external services.
		extSvcCount, err := store.Count(ctx, database.ExternalServicesListOptions{})
		if err != nil {
			return err
		}

		// If we have none configured or we're under the limit, we can pass the
		// validation. Otherwise an error will be returned. Note that we consider
		// a maximum of 0 to be "unlimited", which is consistent with other checks.
		if maxExtSvcCount == 0 || extSvcCount < maxExtSvcCount {
			return nil
		}
		return errcode.NewPresentationError(
			fmt.Sprintf(
				"Unable to create external service: the current plan cannot exceed %d external services (this instance now has %d). Contact Sourcegraph to learn more at https://about.sourcegraph.com/contact/sales.",
				maxExtSvcCount,
				extSvcCount,
			),
		)
	}
}
