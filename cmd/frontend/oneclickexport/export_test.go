package oneclickexport

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExport(t *testing.T) {
	logger := logtest.Scoped(t)

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

	archive, err := exporter.Export(ExportRequest{IncludeSiteConfig: true})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatal(err)
	}

	found := false

	want := `{
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

		if diff := cmp.Diff(want, have); diff != "" {
			t.Fatalf("Exported site config is different. (-want +got):\n%s", diff)
		}
	}

	if !found {
		t.Fatal(errors.New("site config file not found in exported zip archive"))
	}
}
