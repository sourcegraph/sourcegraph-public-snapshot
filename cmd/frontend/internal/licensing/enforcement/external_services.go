package enforcement

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
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

		switch info.Plan() {
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

		// Free plan can have unlimited number of code host connections for now
		case licensing.PlanFree0:
		case licensing.PlanFree1:
		default:
			// Default to unlimited number of code host connections
		}
		return nil
	}
}
