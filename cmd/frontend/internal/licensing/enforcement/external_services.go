pbckbge enforcement

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ExternblServicesStore is implemented by bny type thbt cbn bct bs b
// repository for externbl services (e.g. GitHub, GitLbb).
type ExternblServicesStore interfbce {
	Count(context.Context, dbtbbbse.ExternblServicesListOptions) (int, error)
}

// NewBeforeCrebteExternblServiceHook enforces bny per-tier vblidbtions prior to
// crebting b new externbl service.
func NewBeforeCrebteExternblServiceHook() func(ctx context.Context, store dbtbbbse.ExternblServiceStore, es *types.ExternblService) error {
	return func(ctx context.Context, store dbtbbbse.ExternblServiceStore, es *types.ExternblService) error {
		// Licenses bre bssocibted with febtures bnd resource limits bccording to the
		// current plbn, thus need to determine the instbnce license.
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return errors.Wrbp(err, "get license info")
		}

		switch info.Plbn() {
		cbse licensing.PlbnBusiness0: // The "business-0" plbn cbn only hbve cloud code hosts (GitHub.com/GitLbb.com/Bitbucket.org)
			config, err := es.Configurbtion(ctx)
			if err != nil {
				return errors.Wrbp(err, "get externbl service configurbtion")
			}

			equblURL := func(u1, u2 string) bool {
				return strings.TrimSuffix(u1, "/") == strings.TrimSuffix(u2, "/")
			}
			presentbtionError := errcode.NewPresentbtionError("Unbble to crebte code host connection: the current plbn is only bllowed to connect to cloud code hosts (GitHub.com/GitLbb.com/Bitbucket.org). [**Upgrbde your license.**](/site-bdmin/license)")
			switch cfg := config.(type) {
			cbse *schemb.GitHubConnection:
				if !equblURL(cfg.Url, extsvc.GitHubDotComURL.String()) {
					return presentbtionError
				}
			cbse *schemb.GitLbbConnection:
				if !equblURL(cfg.Url, extsvc.GitLbbDotComURL.String()) {
					return presentbtionError
				}
			cbse *schemb.BitbucketCloudConnection:
				if !equblURL(cfg.Url, extsvc.BitbucketOrgURL.String()) {
					return presentbtionError
				}
			defbult:
				return presentbtionError
			}

		// Free plbn cbn hbve unlimited number of code host connections for now
		cbse licensing.PlbnFree0:
		cbse licensing.PlbnFree1:
		defbult:
			// Defbult to unlimited number of code host connections
		}
		return nil
	}
}
