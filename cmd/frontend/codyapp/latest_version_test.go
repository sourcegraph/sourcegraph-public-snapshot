package codyapp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/updatecheck"
)

func TestLatestVersionHandler(t *testing.T) {
	var resolver = updatecheck.StaticManifestResolver{
		Manifest: updatecheck.AppUpdateManifest{
			Version: "3023.5.8", // set the year part of the version FAR ahead so that there is always a version to update to
			Notes:   "This is a test",
			PubDate: time.Date(2023, time.May, 8, 12, 0, 0, 0, &time.Location{}),
			Platforms: map[string]updatecheck.AppLocation{
				"x86_64-linux": {
					Signature: "Yippy Kay YAY",
					URL:       "https://example.com/linux",
				},
				"x86_64-windows": {
					Signature: "Yippy Kay YAY",
					URL:       "https://example.com/windows",
				},
				"aarch64-darwin": {
					Signature: "Yippy Kay YAY",
					URL:       "https://example.com/darwin",
				},
			},
		},
	}

	var queries = []struct {
		target      string
		arch        string
		expectedURL string
	}{
		{
			"linux",
			"x86_64",
			"https://example.com/linux",
		},
		{
			"windows",
			"x86_64",
			"https://example.com/windows",
		},
		{
			"darwin",
			"aarch64",
			"https://example.com/darwin",
		},
		// if arch and target are empty we provide the release page for the tag
		{
			"",
			"",
			gitHubReleaseBaseURL + resolver.Manifest.GitHubReleaseTag(),
		},
		{
			"toaster",
			"gameboy",
			gitHubReleaseBaseURL + resolver.Manifest.GitHubReleaseTag(),
		},
	}

	for _, q := range queries {
		urlPath := "/app/latest"
		if q.target != "" || q.arch != "" {
			urlPath = fmt.Sprintf("/app/latest?target=%s&arch=%s", q.target, q.arch)
		}

		req := httptest.NewRequest("GET", urlPath, nil)
		w := httptest.NewRecorder()

		latest := newLatestVersion(logtest.NoOp(t), &resolver)
		latest.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected HTTP Status %d for exact version match, but got %d", http.StatusSeeOther, resp.StatusCode)
		}

		loc, err := resp.Location()
		if err != nil {
			t.Fatalf("failed to get location from response: %v", err)
		}

		if loc.String() != q.expectedURL {
			t.Errorf("expected location url %q but got %q", q.expectedURL, loc.String())
		}
	}

}
