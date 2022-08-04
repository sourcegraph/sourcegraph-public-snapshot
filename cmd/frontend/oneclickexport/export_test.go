package oneclickexport

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
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
	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn([]*types.ExternalService{
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test1",
			Config: `{
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
    }`,
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test2",
			Config: `{
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
    }`,
		},
		{
			Kind:        extsvc.KindBitbucketCloud,
			DisplayName: "GitLab - Test1",
			Config: `{
      "url": "https://bitbucket.org",
      "token": "someToken",
      "username": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPathPattern": "bbs/{projectKey}/{repositorySlug}"
    }`,
		},
	}, nil)

	db := database.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

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
	}

	archive, err := exporter.Export(ctx, ExportRequest{IncludeSiteConfig: true, IncludeCodeHostConfig: true})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatal(err)
	}

	wantMap := map[string]string{
		"site-config.json":      wantSiteConfig,
		"code-host-config.json": wantCodeHostConfig,
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

	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn([]*types.ExternalService{
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test1",
			Config: `{
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
    }`,
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test2",
			Config: `{
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
    }`,
		},
		{
			Kind:        extsvc.KindBitbucketCloud,
			DisplayName: "GitLab - Test1",
			Config: `{
      "url": "https://bitbucket.org",
      "token": "someToken",
      "username": "user",
      "repos": [
        "SOURCEGRAPH/repo-0",
        "SOURCEGRAPH/repo-1"
      ],
      "repositoryPathPattern": "bbs/{projectKey}/{repositorySlug}"
    }`,
		},
	}, nil)

	db := database.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

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

	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
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

		fmt.Println(have)

		if diff := cmp.Diff(wantCodeHostConfig, have); diff != "" {
			t.Fatalf("Exported site config is different. (-want +got):\n%s", diff)
		}
	}

	if !found {
		t.Fatal(errors.New("site config file not found in exported zip archive"))
	}
}
