package oneclickexport

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	wantSiteConfig = `{
  "auth.providers": [
    {
      "allowOrgs": [
        "myorg"
      ],
      "clientID": "myclientid",
      "clientSecret": "REDACTED",
      "displayName": "GitHub",
      "type": "github",
      "url": "https://github.com"
    }
  ],
  "experimentalFeatures": {
    "search.index.query.contexts": true
  },
  "permissions.userMapping": {
    "bindID": "username",
    "enabled": true
  }
}`

	wantCodeHostConfig = `[
  {
    "kind": "GITHUB",
    "displayName": "Github - Test1",
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
      "repositoryPathPattern": "github.com/{nameWithOwner}"
    }
  },
  {
    "kind": "GITHUB",
    "displayName": "Github - Test2",
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
      "repositoryPathPattern": "github.com/{nameWithOwner}"
    }
  },
  {
    "kind": "BITBUCKETCLOUD",
    "displayName": "GitLab - Test1",
    "config": {
      "url": "https://bitbucket.org",
      "token": "someToken",
      "username": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPathPattern": "bbs/{projectKey}/{repositorySlug}"
    }
  }
]`
	wantExtSvcDBQueryResult = `[
  {
    "ID": 0,
    "Kind": "GITHUB",
    "DisplayName": "Github - Test1",
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
      "repositoryPathPattern": "github.com/{nameWithOwner}"
    },
    "CreatedAt": "0001-01-01T00:00:00Z",
    "UpdatedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": "0001-01-01T00:00:00Z",
    "LastSyncAt": "0001-01-01T00:00:00Z",
    "NextSyncAt": "0001-01-01T00:00:00Z",
    "Unrestricted": false,
    "HasWebhooks": null,
    "TokenExpiresAt": null
  },
  {
    "ID": 0,
    "Kind": "GITHUB",
    "DisplayName": "Github - Test2",
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
      "repositoryPathPattern": "github.com/{nameWithOwner}"
    },
    "CreatedAt": "0001-01-01T00:00:00Z",
    "UpdatedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": "0001-01-01T00:00:00Z",
    "LastSyncAt": "0001-01-01T00:00:00Z",
    "NextSyncAt": "0001-01-01T00:00:00Z",
    "Unrestricted": false,
    "HasWebhooks": null,
    "TokenExpiresAt": null
  },
  {
    "ID": 0,
    "Kind": "BITBUCKETCLOUD",
    "DisplayName": "GitLab - Test1",
    "Config": {
      "url": "https://bitbucket.org",
      "token": "someToken",
      "username": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPathPattern": "bbs/{projectKey}/{repositorySlug}"
    },
    "CreatedAt": "0001-01-01T00:00:00Z",
    "UpdatedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": "0001-01-01T00:00:00Z",
    "LastSyncAt": "0001-01-01T00:00:00Z",
    "NextSyncAt": "0001-01-01T00:00:00Z",
    "Unrestricted": false,
    "HasWebhooks": null,
    "TokenExpiresAt": null
  }
]`
	wantExtSvcReposDBQueryResult = `[
  {
    "externalServiceID": 1,
    "repoID": 1,
    "cloneURL": "cloneUrl",
    "createdAt": "0001-01-01T00:00:00Z"
  },
  {
    "externalServiceID": 1,
    "repoID": 2,
    "cloneURL": "cloneUrl",
    "createdAt": "0001-01-01T00:00:00Z"
  }
]`
)

func TestExport(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthProviders: []schema.AuthProviders{{
				Github: &schema.GitHubAuthProvider{
					ClientID:     "myclientid",
					ClientSecret: "myclientsecret",
					DisplayName:  "GitHub",
					Type:         extsvc.TypeGitHub,
					Url:          "https://github.com",
					AllowOrgs:    []string{"myorg"},
				},
			}},
			PermissionsUserMapping: &schema.PermissionsUserMapping{
				BindID:  "username",
				Enabled: true,
			},
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				SearchIndexQueryContexts: true,
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	exporter := &DataExporter{
		logger: logger,
		configProcessors: map[string]Processor[ConfigRequest]{
			"siteConfig": &SiteConfigProcessor{
				logger: logger,
				Type:   "siteConfig",
			},
		},
	}

	archive, err := exporter.Export(ctx, ExportRequest{IncludeSiteConfig: true})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := getArchiveReader(archive)
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, f := range zr.File {
		if f.Name != "site-config.json" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		haveBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}

		have := string(haveBytes)

		if diff := cmp.Diff(wantSiteConfig, have); diff != "" {
			t.Fatalf("Exported site config is different. (-want +got):\n%s", diff)
		}
	}

	if !found {
		t.Fatal(errors.New("site config file not found in exported zip archive"))
	}
}

// TestExport_CumulativeTest is a test with a full export available at the
// moment. This test should be updated with every new added piece of exported
// data.
func TestExport_CumulativeTest(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	// Mocking site config
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthProviders: []schema.AuthProviders{{
				Github: &schema.GitHubAuthProvider{
					ClientID:     "myclientid",
					ClientSecret: "myclientsecret",
					DisplayName:  "GitHub",
					Type:         extsvc.TypeGitHub,
					Url:          "https://github.com",
					AllowOrgs:    []string{"myorg"},
				},
			}},
			PermissionsUserMapping: &schema.PermissionsUserMapping{
				BindID:  "username",
				Enabled: true,
			},
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				SearchIndexQueryContexts: true,
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })

	// Mocking external services for code host configs export
	db := mockExternalServicesDB()

	exporter := &DataExporter{
		logger: logger,
		configProcessors: map[string]Processor[ConfigRequest]{
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
		dbProcessors: map[string]Processor[Limit]{
			"externalServices": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externalServices",
			},
		},
	}

	archive, err := exporter.Export(ctx, ExportRequest{
		IncludeSiteConfig:     true,
		IncludeCodeHostConfig: true,
		DBQueries: []*DBQueryRequest{{
			TableName: "externalServices",
			Count:     1,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := getArchiveReader(archive)
	if err != nil {
		t.Fatal(err)
	}

	wantMap := map[string]string{
		"site-config.json":         wantSiteConfig,
		"code-host-config.json":    wantCodeHostConfig,
		"db-external-services.txt": wantExtSvcDBQueryResult,
	}

	for _, f := range zr.File {
		currentFileName := f.Name
		want, ok := wantMap[currentFileName]
		if !ok {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		haveBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}

		delete(wantMap, currentFileName)

		have := string(haveBytes)

		if diff := cmp.Diff(want, have); diff != "" {
			t.Errorf("%q has wrong content. (-want +got):\n%s", currentFileName, diff)
		}
	}

	for file := range wantMap {
		t.Errorf("Missing file from ZIP archive: %q", file)
	}
}

func TestExport_CodeHostConfigs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := mockExternalServicesDB()

	exporter := &DataExporter{
		logger: logger,
		configProcessors: map[string]Processor[ConfigRequest]{
			"codeHostConfig": &CodeHostConfigProcessor{
				db:     db,
				logger: logger,
				Type:   "codeHostConfig",
			},
		},
	}

	archive, err := exporter.Export(ctx, ExportRequest{IncludeCodeHostConfig: true})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := getArchiveReader(archive)
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, f := range zr.File {
		if f.Name != "code-host-config.json" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		haveBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}

		have := string(haveBytes)

		if diff := cmp.Diff(wantCodeHostConfig, have); diff != "" {
			t.Fatalf("Exported site config is different. (-want +got):\n%s", diff)
		}
	}

	if !found {
		t.Fatal(errors.New("site config file not found in exported zip archive"))
	}
}

func mockExternalServicesDB() *dbmocks.MockDB {
	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn([]*types.ExternalService{
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test1",
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
      "repositoryPathPattern": "github.com/{nameWithOwner}"
    }`),
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test2",
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
      "repositoryPathPattern": "github.com/{nameWithOwner}"
    }`),
		},
		{
			Kind:        extsvc.KindBitbucketCloud,
			DisplayName: "GitLab - Test1",
			Config: extsvc.NewUnencryptedConfig(`{
      "url": "https://bitbucket.org",
      "token": "someToken",
      "username": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPathPattern": "bbs/{projectKey}/{repositorySlug}"
    }`),
		},
	}, nil)

	db := dbmocks.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	return db
}

func TestExport_DB_ExternalServices(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := mockExternalServicesDB()

	exporter := &DataExporter{
		logger: logger,
		dbProcessors: map[string]Processor[Limit]{
			"externalServices": ExtSvcQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externalServices",
			},
		},
	}

	archive, err := exporter.Export(ctx, ExportRequest{DBQueries: []*DBQueryRequest{{
		TableName: "externalServices",
		Count:     1,
	}}})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := getArchiveReader(archive)
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, f := range zr.File {
		if f.Name != "db-external-services.txt" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		haveBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}

		have := string(haveBytes)

		if diff := cmp.Diff(wantExtSvcDBQueryResult, have); diff != "" {
			t.Fatalf("Exported external services are different. (-want +got):\n%s", diff)
		}
	}

	if !found {
		t.Fatal(errors.New("external services file not found in exported zip archive"))
	}
}

func TestExport_DB_ExternalServiceRepos(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListReposFunc.SetDefaultReturn([]*types.ExternalServiceRepo{
		{
			ExternalServiceID: 1,
			RepoID:            1,
			CloneURL:          "cloneUrl",
			CreatedAt:         time.Time{},
		},
		{
			ExternalServiceID: 1,
			RepoID:            2,
			CloneURL:          "cloneUrl",
			CreatedAt:         time.Time{},
		},
	},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	exporter := &DataExporter{
		logger: logger,
		dbProcessors: map[string]Processor[Limit]{
			"externalServiceRepos": ExtSvcReposQueryProcessor{
				db:     db,
				logger: logger,
				Type:   "externalServiceRepos",
			},
		},
	}

	archive, err := exporter.Export(ctx, ExportRequest{DBQueries: []*DBQueryRequest{{
		TableName: "externalServiceRepos",
		Count:     2,
	}}})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := getArchiveReader(archive)
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, f := range zr.File {
		if f.Name != "db-external-service-repos.txt" {
			continue
		}
		found = true
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		haveBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}

		have := string(haveBytes)

		if diff := cmp.Diff(wantExtSvcReposDBQueryResult, have); diff != "" {
			t.Fatalf("Exported external services are different. (-want +got):\n%s", diff)
		}
	}

	if !found {
		t.Fatal(errors.New("external services file not found in exported zip archive"))
	}
}

func getArchiveReader(archive io.Reader) (*zip.Reader, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(archive)
	if err != nil {
		return nil, err
	}
	return zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
}
