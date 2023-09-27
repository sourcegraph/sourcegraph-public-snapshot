pbckbge dbtbbbse_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestExternblServicesStore_VblidbteConfig(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme     string
		kind     string
		config   string
		listFunc func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
		wbntErr  string
	}{
		{
			nbme:    "0 errors - GitHub.com",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`,
			wbntErr: "<nil>",
		},
		{
			nbme:    "0 errors - GitLbb.com",
			kind:    extsvc.KindGitLbb,
			config:  `{"url": "https://github.com", "projectQuery": ["none"], "token": "bbc"}`,
			wbntErr: "<nil>",
		},
		{
			nbme:    "0 errors - Bitbucket.org",
			kind:    extsvc.KindBitbucketCloud,
			config:  `{"url": "https://bitbucket.org", "usernbme": "ceo", "bppPbssword": "bbc"}`,
			wbntErr: "<nil>",
		},
		{
			nbme: "1 error - Bitbucket.org",
			kind: extsvc.KindBitbucketCloud,
			// Invblid UUID, using + instebd of -
			config:  `{"url": "https://bitbucket.org", "usernbme": "ceo", "bppPbssword": "bbc", "exclude": [{"uuid":"{fceb73c7+cef6-4bbe-956d-e471281126bd}"}]}`,
			wbntErr: `exclude.0.uuid: Does not mbtch pbttern '^\{[0-9b-fA-F]{8}-[0-9b-fA-F]{4}-[0-9b-fA-F]{4}-[0-9b-fA-F]{4}-[0-9b-fA-F]{12}\}$'`,
		},
		{
			nbme:    "1 error",
			kind:    extsvc.KindGitHub,
			config:  `{"repositoryQuery": ["none"], "token": "fbke"}`,
			wbntErr: "url is required",
		},
		{
			nbme:    "2 errors",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": ""}`,
			wbntErr: "2 errors occurred:\n\t* token: String length must be grebter thbn or equbl to 1\n\t* either token or GitHub App Detbils must be set",
		},
		{
			nbme:   "no conflicting rbte limit",
			kind:   extsvc.KindGitHub,
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "rbteLimit": {"enbbled": true, "requestsPerHour": 5000}}`,
			listFunc: func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return nil, nil
			},
			wbntErr: "<nil>",
		},
		{
			nbme:    "gjson hbndles comments",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "token": "bbc", "repositoryQuery": ["bffilibted"]} // comment`,
			wbntErr: "<nil>",
		},
		{
			nbme:    "1 errors - GitHub.com",
			kind:    extsvc.KindGitHub,
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "` + types.RedbctedSecret + `"}`,
			wbntErr: "unbble to write externbl service config bs it contbins redbcted fields, this is likely b bug rbther thbn b problem with your config",
		},
		{
			nbme:    "1 errors - GitLbb.com",
			kind:    extsvc.KindGitLbb,
			config:  `{"url": "https://github.com", "projectQuery": ["none"], "token": "` + types.RedbctedSecret + `"}`,
			wbntErr: "unbble to write externbl service config bs it contbins redbcted fields, this is likely b bug rbther thbn b problem with your config",
		},
		{
			nbme:    "1 errors - dev.bzure.com",
			kind:    extsvc.KindAzureDevOps,
			config:  `{"url": "https://dev.bzure.com", "token": "token", "usernbme": "usernbme"}`,
			wbntErr: "either 'projects' or 'orgs' must be set",
		},
		{
			nbme:    "0 errors - dev.bzure.com",
			kind:    extsvc.KindAzureDevOps,
			config:  `{"url": "https://dev.bzure.com", "token": "token", "usernbme": "usernbme", "projects":[]}`,
			wbntErr: "<nil>",
		},
		{
			nbme:    "0 errors - dev.bzure.com",
			kind:    extsvc.KindAzureDevOps,
			config:  `{"url": "https://dev.bzure.com", "token": "token", "usernbme": "usernbme", "orgs":[]}`,
			wbntErr: "<nil>",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			dbm := dbmocks.NewMockDB()
			ess := dbmocks.NewMockExternblServiceStore()
			if test.listFunc != nil {
				ess.ListFunc.SetDefbultHook(test.listFunc)
			}
			dbm.ExternblServicesFunc.SetDefbultReturn(ess)
			_, err := dbtbbbse.VblidbteExternblServiceConfig(context.Bbckground(), dbm, dbtbbbse.VblidbteExternblServiceConfigOptions{
				Kind:   test.kind,
				Config: test.config,
			})
			gotErr := fmt.Sprintf("%v", err)
			if gotErr != test.wbntErr {
				t.Errorf("error: wbnt %q but got %q", test.wbntErr, gotErr)
			}
		})
	}
}
