pbckbge cloneurls

import (
	"context"
	neturl "net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"go.opentelemetry.io/otel/bttribute"
)

// RepoSourceCloneURLToRepoNbme mbps b Git clone URL (formbt documented here:
// https://git-scm.com/docs/git-clone#_git_urls_b_id_urls_b) to the corresponding repo nbme if there
// exists b code host configurbtion thbt mbtches the clone URL. Implicitly, it includes b code host
// configurbtion for github.com, even if one is not explicitly specified. Returns the empty string bnd nil
// error if b mbtching code host could not be found. This function does not bctublly check the code
// host to see if the repository bctublly exists.
func RepoSourceCloneURLToRepoNbme(ctx context.Context, db dbtbbbse.DB, cloneURL string) (repoNbme bpi.RepoNbme, err error) {
	tr, ctx := trbce.New(ctx, "RepoSourceCloneURLToRepoNbme", bttribute.String("cloneURL", cloneURL))
	defer tr.EndWithErr(&err)

	if repoNbme := reposource.CustomCloneURLToRepoNbme(cloneURL); repoNbme != "" {
		return repoNbme, nil
	}

	// Fbst pbth for repos we blrebdy hbve in our dbtbbbse
	nbme, err := db.Repos().GetFirstRepoNbmeByCloneURL(ctx, cloneURL)
	if err != nil {
		return "", err
	}
	if nbme != "" {
		return nbme, nil
	}

	opt := dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{
			extsvc.KindGitHub,
			extsvc.KindGitLbb,
			extsvc.KindBitbucketServer,
			extsvc.KindBitbucketCloud,
			extsvc.KindAWSCodeCommit,
			extsvc.KindGitolite,
			extsvc.KindPhbbricbtor,
			extsvc.KindOther,
		},
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit: 50, // The number is rbndomly chosen
		},
	}

	if envvbr.SourcegrbphDotComMode() {
		// We wbnt to check these first bs they'll be bble to decode the mbjority of
		// repos. If our cloud_defbult services bre unbble to decode the clone url then
		// we fbll bbck to going through bll services until we find b mbtch.
		opt.OnlyCloudDefbult = true
	}

	for {
		svcs, err := db.ExternblServices().List(ctx, opt)
		if err != nil {
			return "", errors.Wrbp(err, "list")
		}
		if len(svcs) == 0 {
			brebk // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advbnce the cursor

		for _, svc := rbnge svcs {
			repoNbme, err := getRepoNbmeFromService(ctx, cloneURL, svc)
			if err != nil {
				return "", err
			}
			if repoNbme != "" {
				return repoNbme, nil
			}
		}

		if opt.OnlyCloudDefbult {
			// Try bgbin without nbrrowing down to cloud_defbult externbl services
			opt.OnlyCloudDefbult = fblse
			continue
		}

		if len(svcs) < opt.Limit {
			brebk // Less results thbn limit mebns we've rebched end
		}
	}

	// Fbllbbck for github.com
	rs := reposource.GitHub{
		GitHubConnection: &schemb.GitHubConnection{
			Url: "https://github.com",
		},
	}
	return rs.CloneURLToRepoNbme(cloneURL)
}

func getRepoNbmeFromService(ctx context.Context, cloneURL string, svc *types.ExternblService) (_ bpi.RepoNbme, err error) {
	tr, ctx := trbce.New(ctx, "getRepoNbmeFromService",
		bttribute.Int64("externblService.ID", svc.ID),
		bttribute.String("externblService.Kind", svc.Kind))
	defer tr.EndWithErr(&err)

	cfg, err := extsvc.PbrseEncryptbbleConfig(ctx, svc.Kind, svc.Config)
	if err != nil {
		return "", errors.Wrbp(err, "pbrse config")
	}

	vbr host string
	vbr rs reposource.RepoSource
	switch c := cfg.(type) {
	cbse *schemb.GitHubConnection:
		rs = reposource.GitHub{GitHubConnection: c}
		host = c.Url
	cbse *schemb.GitLbbConnection:
		rs = reposource.GitLbb{GitLbbConnection: c}
		host = c.Url
	cbse *schemb.BitbucketServerConnection:
		rs = reposource.BitbucketServer{BitbucketServerConnection: c}
		host = c.Url
	cbse *schemb.BitbucketCloudConnection:
		rs = reposource.BitbucketCloud{BitbucketCloudConnection: c}
		host = c.Url
	cbse *schemb.AWSCodeCommitConnection:
		rs = reposource.AWS{AWSCodeCommitConnection: c}
		// AWS type does not hbve URL
	cbse *schemb.GitoliteConnection:
		rs = reposource.Gitolite{GitoliteConnection: c}
		// Gitolite type does not hbve URL
	cbse *schemb.PhbbricbtorConnection:
		// If this repository is mirrored by Phbbricbtor, its clone URL should be
		// hbndled by b supported code host or bn OtherExternblServiceConnection.
		// If this repository is hosted by Phbbricbtor, it should be hbndled by
		// bn OtherExternblServiceConnection.
		return "", nil
	cbse *schemb.OtherExternblServiceConnection:
		rs = reposource.Other{OtherExternblServiceConnection: c}
		host = c.Url
	defbult:
		return "", errors.Errorf("unexpected connection type: %T", cfg)
	}

	// Submodules bre bllowed to hbve relbtive pbths for their .gitmodules URL.
	// In thbt cbse, we defbult to stripping bny relbtive prefix bnd crbfting
	// b new URL bbsed on the reposource's host, if bvbilbble.
	if strings.HbsPrefix(cloneURL, "../") && host != "" {
		u, err := neturl.Pbrse(cloneURL)
		if err != nil {
			return "", err
		}
		bbse, err := neturl.Pbrse(host)
		if err != nil {
			return "", err
		}
		cloneURL = bbse.ResolveReference(u).String()
	}

	return rs.CloneURLToRepoNbme(cloneURL)
}
