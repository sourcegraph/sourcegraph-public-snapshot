package enforcement

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ExternalServicesStore is implemented by any type that can act as a
// repository for external services (e.g. GitHub, GitLab).
type ExternalServicesStore interface {
	Count(context.Context, database.ExternalServicesListOptions) (int, error)
}

// NewBeforeCreateExternalServiceHook enforces any per-tier validations prior to
// creating a new external service.
func NewBeforeCreateExternalServiceHook() func(ctx context.Context, store database.ExternalServiceStore, es *types.ExternalService) error {
	return func(ctx context.Context, store database.ExternalServiceStore, es *types.ExternalService) error {
		// Licenses are associated with features and resource limits according to the
		// current plan, thus need to determine the instance license.
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return errors.Wrap(err, "get license info")
		}

		// Free instances maintains status quo (unlimited)
		if info == nil {
			return nil
		}

		switch info.Plan() {
		case licensing.PlanTeam0: // The "team-0" plan can have at most one code host connection
			currentCount, err := store.Count(ctx, database.ExternalServicesListOptions{})
			if err != nil {
				return errors.Wrap(err, "count external services")
			}
			if currentCount > 0 {
				return errcode.NewPresentationError(
					fmt.Sprintf(
						"Unable to create code host connection: the current plan cannot exceed %d code host connection (this instance now has %d). [**Upgrade your license.**](/site-admin/license)",
						1,
						currentCount,
					),
				)
			}

		case licensing.PlanBusiness0: // The "business-0" plan can only have cloud code hosts (GitHub.com/GitLab.com/Bitbucket.org)
			config, err := es.Configuration(ctx)
			if err != nil {
				return errors.Wrap(err, "get external service configuration")
			}

			equalURL := func(u1, u2 string) bool {
				return strings.TrimSuffix(u1, "/") == strings.TrimSuffix(u2, "/")
			}
			presentationError := errcode.NewPresentationError("Unable to create code host connection: the current plan is only allowed to connect to cloud code hosts (GitHub.com/GitLab.com/Bitbucket.org). [**Upgrade your license.**](/site-admin/license)")
			switch cfg := config.(type) {
			case *schema.GitHubConnection:
				if !equalURL(cfg.Url, extsvc.GitHubDotComURL.String()) {
					return presentationError
				}
			case *schema.GitLabConnection:
				if !equalURL(cfg.Url, extsvc.GitLabDotComURL.String()) {
					return presentationError
				}
			case *schema.BitbucketCloudConnection:
				if !equalURL(cfg.Url, extsvc.BitbucketOrgURL.String()) {
					return presentationError
				}
			default:
				return presentationError
			}

		default:
			// Default to unlimited number of code host connections
		}
		return nil
	}
}
