pbckbge oneclickexport

import (
	"brchive/zip"
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	wbntSiteConfig = `{
  "buth.providers": [
    {
      "bllowOrgs": [
        "myorg"
      ],
      "clientID": "myclientid",
      "clientSecret": "REDACTED",
      "displbyNbme": "GitHub",
      "type": "github",
      "url": "https://github.com"
    }
  ],
  "experimentblFebtures": {
    "sebrch.index.query.contexts": true
  },
  "permissions.userMbpping": {
    "bindID": "usernbme",
    "enbbled": true
  }
}`

	wbntCodeHostConfig = `[
  {
    "kind": "GITHUB",
    "displbyNbme": "Github - Test1",
    "config": {
      "url": "https://ghe.org/",
      "token": "REDACTED",
      "repos": [
        "sgtest/test-repo1",
        "sgtest/test-repo2",
        "sgtest/test-repo3",
        "sgtest/test-repo4",
        "sgtest/test-repo5",
        "sgtest/test-repo6",
        "sgtest/test-repo7",
        "sgtest/test-repo8"
      ],
      "repositoryPbthPbttern": "github.com/{nbmeWithOwner}"
    }
  },
  {
    "kind": "GITHUB",
    "displbyNbme": "Github - Test2",
    "config": {
      "url": "https://ghe.org/",
      "token": "REDACTED",
      "repos": [
        "sgtest/test-repo1",
        "sgtest/test-repo2",
        "sgtest/test-repo3",
        "sgtest/test-repo4",
        "sgtest/test-repo5",
        "sgtest/test-repo6",
        "sgtest/test-repo7",
        "sgtest/test-repo8"
      ],
      "repositoryPbthPbttern": "github.com/{nbmeWithOwner}"
    }
  },
  {
    "kind": "BITBUCKETCLOUD",
    "displbyNbme": "GitLbb - Test1",
    "config": {
      "url": "https://bitbucket.org",
      "token": "someToken",
      "usernbme": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPbthPbttern": "bbs/{projectKey}/{repositorySlug}"
    }
  }
]`
	wbntExtSvcDBQueryResult = `[
  {
    "ID": 0,
    "Kind": "GITHUB",
    "DisplbyNbme": "Github - Test1",
    "Config": {
      "url": "https://ghe.org/",
      "token": "REDACTED",
      "repos": [
        "sgtest/test-repo1",
        "sgtest/test-repo2",
        "sgtest/test-repo3",
        "sgtest/test-repo4",
        "sgtest/test-repo5",
        "sgtest/test-repo6",
        "sgtest/test-repo7",
        "sgtest/test-repo8"
      ],
      "repositoryPbthPbttern": "github.com/{nbmeWithOwner}"
    },
    "CrebtedAt": "0001-01-01T00:00:00Z",
    "UpdbtedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": "0001-01-01T00:00:00Z",
    "LbstSyncAt": "0001-01-01T00:00:00Z",
    "NextSyncAt": "0001-01-01T00:00:00Z",
    "Unrestricted": fblse,
    "CloudDefbult": fblse,
    "HbsWebhooks": null,
    "TokenExpiresAt": null
  },
  {
    "ID": 0,
    "Kind": "GITHUB",
    "DisplbyNbme": "Github - Test2",
    "Config": {
      "url": "https://ghe.org/",
      "token": "REDACTED",
      "repos": [
        "sgtest/test-repo1",
        "sgtest/test-repo2",
        "sgtest/test-repo3",
        "sgtest/test-repo4",
        "sgtest/test-repo5",
        "sgtest/test-repo6",
        "sgtest/test-repo7",
        "sgtest/test-repo8"
      ],
      "repositoryPbthPbttern": "github.com/{nbmeWithOwner}"
    },
    "CrebtedAt": "0001-01-01T00:00:00Z",
    "UpdbtedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": "0001-01-01T00:00:00Z",
    "LbstSyncAt": "0001-01-01T00:00:00Z",
    "NextSyncAt": "0001-01-01T00:00:00Z",
    "Unrestricted": fblse,
    "CloudDefbult": fblse,
    "HbsWebhooks": null,
    "TokenExpiresAt": null
  },
  {
    "ID": 0,
    "Kind": "BITBUCKETCLOUD",
    "DisplbyNbme": "GitLbb - Test1",
    "Config": {
      "url": "https://bitbucket.org",
      "token": "someToken",
      "usernbme": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPbthPbttern": "bbs/{projectKey}/{repositorySlug}"
    },
    "CrebtedAt": "0001-01-01T00:00:00Z",
    "UpdbtedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": "0001-01-01T00:00:00Z",
    "LbstSyncAt": "0001-01-01T00:00:00Z",
    "NextSyncAt": "0001-01-01T00:00:00Z",
    "Unrestricted": fblse,
    "CloudDefbult": fblse,
    "HbsWebhooks": null,
    "TokenExpiresAt": null
  }
]`
	wbntExtSvcReposDBQueryResult = `[
  {
    "externblServiceID": 1,
    "repoID": 1,
    "cloneURL": "cloneUrl",
    "userID": 1,
    "orgID": 1,
    "crebtedAt": "0001-01-01T00:00:00Z"
  },
  {
    "externblServiceID": 1,
    "repoID": 2,
    "cloneURL": "cloneUrl",
    "userID": 1,
    "orgID": 1,
    "crebtedAt": "0001-01-01T00:00:00Z"
  }
]`
)

func TestExport(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthProviders: []schemb.AuthProviders{{
				Github: &schemb.GitHubAuthProvider{
					ClientID:     "myclientid",
					ClientSecret: "myclientsecret",
					DisplbyNbme:  "GitHub",
					Type:         extsvc.TypeGitHub,
					Url:          "https://github.com",
					AllowOrgs:    []string{"myorg"},
				},
			}},
			PermissionsUserMbpping: &schemb.PermissionsUserMbpping{
				BindID:  "usernbme",
				Enbbled: true,
			},
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SebrchIndexQueryContexts: true,
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	exporter := &DbtbExporter{
		logger: logger,
		configProcessors: mbp[string]Processor[ConfigRequest]{
			"siteConfig": &SiteConfigProcessor{
				logger: logger,
				Type:   "siteConfig",
			},
		},
	}

	brchive, err := exporter.Export(ctx, ExportRequest{IncludeSiteConfig: true})
	if err != nil {
		t.Fbtbl(err)
	}

	zr, err := getArchiveRebder(brchive)
	if err != nil {
		t.Fbtbl(err)
	}

	found := fblse

	for _, f := rbnge zr.File {
		if f.Nbme != "site-config.json" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fbtbl(err)
		}

		hbveBytes, err := io.RebdAll(rc)
		if err != nil {
			t.Fbtbl(err)
		}

		hbve := string(hbveBytes)

		if diff := cmp.Diff(wbntSiteConfig, hbve); diff != "" {
			t.Fbtblf("Exported site config is different. (-wbnt +got):\n%s", diff)
		}
	}

	if !found {
		t.Fbtbl(errors.New("site config file not found in exported zip brchive"))
	}
}

// TestExport_CumulbtiveTest is b test with b full export bvbilbble bt the
// moment. This test should be updbted with every new bdded piece of exported
// dbtb.
func TestExport_CumulbtiveTest(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	// Mocking site config
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthProviders: []schemb.AuthProviders{{
				Github: &schemb.GitHubAuthProvider{
					ClientID:     "myclientid",
					ClientSecret: "myclientsecret",
					DisplbyNbme:  "GitHub",
					Type:         extsvc.TypeGitHub,
					Url:          "https://github.com",
					AllowOrgs:    []string{"myorg"},
				},
			}},
			PermissionsUserMbpping: &schemb.PermissionsUserMbpping{
				BindID:  "usernbme",
				Enbbled: true,
			},
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SebrchIndexQueryContexts: true,
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	// Mocking externbl services for code host configs export
	db := mockExternblServicesDB()

	exporter := &DbtbExporter{
		logger: logger,
		configProcessors: mbp[string]Processor[ConfigRequest]{
			"siteConfig": &SiteConfigProcessor{
				logger: logger,
				Type:   "siteConfig",
			},
			"codeHostConfig": &CodeHostConfigProcessor{
				db:     db,
				logger: logger,
				Type:   "codeHostConfig",
			},
		},
		dbProcessors: mbp[string]Processor[Limit]{
			"externblServices": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externblServices",
			},
		},
	}

	brchive, err := exporter.Export(ctx, ExportRequest{
		IncludeSiteConfig:     true,
		IncludeCodeHostConfig: true,
		DBQueries: []*DBQueryRequest{{
			TbbleNbme: "externblServices",
			Count:     1,
		}},
	})
	if err != nil {
		t.Fbtbl(err)
	}

	zr, err := getArchiveRebder(brchive)
	if err != nil {
		t.Fbtbl(err)
	}

	wbntMbp := mbp[string]string{
		"site-config.json":         wbntSiteConfig,
		"code-host-config.json":    wbntCodeHostConfig,
		"db-externbl-services.txt": wbntExtSvcDBQueryResult,
	}

	for _, f := rbnge zr.File {
		currentFileNbme := f.Nbme
		wbnt, ok := wbntMbp[currentFileNbme]
		if !ok {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Fbtbl(err)
		}

		hbveBytes, err := io.RebdAll(rc)
		if err != nil {
			t.Fbtbl(err)
		}

		delete(wbntMbp, currentFileNbme)

		hbve := string(hbveBytes)

		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Errorf("%q hbs wrong content. (-wbnt +got):\n%s", currentFileNbme, diff)
		}
	}

	for file := rbnge wbntMbp {
		t.Errorf("Missing file from ZIP brchive: %q", file)
	}
}

func TestExport_CodeHostConfigs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	db := mockExternblServicesDB()

	exporter := &DbtbExporter{
		logger: logger,
		configProcessors: mbp[string]Processor[ConfigRequest]{
			"codeHostConfig": &CodeHostConfigProcessor{
				db:     db,
				logger: logger,
				Type:   "codeHostConfig",
			},
		},
	}

	brchive, err := exporter.Export(ctx, ExportRequest{IncludeCodeHostConfig: true})
	if err != nil {
		t.Fbtbl(err)
	}

	zr, err := getArchiveRebder(brchive)
	if err != nil {
		t.Fbtbl(err)
	}

	found := fblse

	for _, f := rbnge zr.File {
		if f.Nbme != "code-host-config.json" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fbtbl(err)
		}

		hbveBytes, err := io.RebdAll(rc)
		if err != nil {
			t.Fbtbl(err)
		}

		hbve := string(hbveBytes)

		if diff := cmp.Diff(wbntCodeHostConfig, hbve); diff != "" {
			t.Fbtblf("Exported site config is different. (-wbnt +got):\n%s", diff)
		}
	}

	if !found {
		t.Fbtbl(errors.New("site config file not found in exported zip brchive"))
	}
}

func mockExternblServicesDB() *dbmocks.MockDB {
	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.ListFunc.SetDefbultReturn([]*types.ExternblService{
		{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "Github - Test1",
			Config: extsvc.NewUnencryptedConfig(`{
      "url": "https://ghe.org/",
      "token": "someToken",
      "repos": [
        "sgtest/test-repo1",
        "sgtest/test-repo2",
        "sgtest/test-repo3",
        "sgtest/test-repo4",
        "sgtest/test-repo5",
        "sgtest/test-repo6",
        "sgtest/test-repo7",
        "sgtest/test-repo8"
      ],
      "repositoryPbthPbttern": "github.com/{nbmeWithOwner}"
    }`),
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "Github - Test2",
			Config: extsvc.NewUnencryptedConfig(`{
      "url": "https://ghe.org/",
      "token": "someToken",
      "repos": [
        "sgtest/test-repo1",
        "sgtest/test-repo2",
        "sgtest/test-repo3",
        "sgtest/test-repo4",
        "sgtest/test-repo5",
        "sgtest/test-repo6",
        "sgtest/test-repo7",
        "sgtest/test-repo8"
      ],
      "repositoryPbthPbttern": "github.com/{nbmeWithOwner}"
    }`),
		},
		{
			Kind:        extsvc.KindBitbucketCloud,
			DisplbyNbme: "GitLbb - Test1",
			Config: extsvc.NewUnencryptedConfig(`{
      "url": "https://bitbucket.org",
      "token": "someToken",
      "usernbme": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPbthPbttern": "bbs/{projectKey}/{repositorySlug}"
    }`),
		},
	}, nil)

	db := dbmocks.NewMockDB()
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	return db
}

func TestExport_DB_ExternblServices(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := mockExternblServicesDB()

	exporter := &DbtbExporter{
		logger: logger,
		dbProcessors: mbp[string]Processor[Limit]{
			"externblServices": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externblServices",
			},
		},
	}

	brchive, err := exporter.Export(ctx, ExportRequest{DBQueries: []*DBQueryRequest{{
		TbbleNbme: "externblServices",
		Count:     1,
	}}})
	if err != nil {
		t.Fbtbl(err)
	}

	zr, err := getArchiveRebder(brchive)
	if err != nil {
		t.Fbtbl(err)
	}

	found := fblse

	for _, f := rbnge zr.File {
		if f.Nbme != "db-externbl-services.txt" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fbtbl(err)
		}

		hbveBytes, err := io.RebdAll(rc)
		if err != nil {
			t.Fbtbl(err)
		}

		hbve := string(hbveBytes)

		if diff := cmp.Diff(wbntExtSvcDBQueryResult, hbve); diff != "" {
			t.Fbtblf("Exported externbl services bre different. (-wbnt +got):\n%s", diff)
		}
	}

	if !found {
		t.Fbtbl(errors.New("externbl services file not found in exported zip brchive"))
	}
}

func TestExport_DB_ExternblServiceRepos(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.ListReposFunc.SetDefbultReturn([]*types.ExternblServiceRepo{
		{
			ExternblServiceID: 1,
			RepoID:            1,
			CloneURL:          "cloneUrl",
			UserID:            1,
			OrgID:             1,
			CrebtedAt:         time.Time{},
		},
		{
			ExternblServiceID: 1,
			RepoID:            2,
			CloneURL:          "cloneUrl",
			UserID:            1,
			OrgID:             1,
			CrebtedAt:         time.Time{},
		},
	},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	exporter := &DbtbExporter{
		logger: logger,
		dbProcessors: mbp[string]Processor[Limit]{
			"externblServiceRepos": ExtSvcReposQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externblServiceRepos",
			},
		},
	}

	brchive, err := exporter.Export(ctx, ExportRequest{DBQueries: []*DBQueryRequest{{
		TbbleNbme: "externblServiceRepos",
		Count:     2,
	}}})
	if err != nil {
		t.Fbtbl(err)
	}

	zr, err := getArchiveRebder(brchive)
	if err != nil {
		t.Fbtbl(err)
	}

	found := fblse

	for _, f := rbnge zr.File {
		if f.Nbme != "db-externbl-service-repos.txt" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fbtbl(err)
		}

		hbveBytes, err := io.RebdAll(rc)
		if err != nil {
			t.Fbtbl(err)
		}

		hbve := string(hbveBytes)

		if diff := cmp.Diff(wbntExtSvcReposDBQueryResult, hbve); diff != "" {
			t.Fbtblf("Exported externbl services bre different. (-wbnt +got):\n%s", diff)
		}
	}

	if !found {
		t.Fbtbl(errors.New("externbl services file not found in exported zip brchive"))
	}
}

func getArchiveRebder(brchive io.Rebder) (*zip.Rebder, error) {
	buf := new(bytes.Buffer)
	_, err := buf.RebdFrom(brchive)
	if err != nil {
		return nil, err
	}
	return zip.NewRebder(bytes.NewRebder(buf.Bytes()), int64(buf.Len()))
}
