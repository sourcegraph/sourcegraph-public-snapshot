package codyapp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
)

func TestLatestVersionHandler(t *testing.T) {
	var resolver = StaticManifestResolver{
		Manifest: AppUpdateManifest{
			Version: "3023.5.8", // set the year part of the version FAR ahead so that there is always a version to update to
			Notes:   "This is a test",
			PubDate: time.Date(2023, time.May, 8, 12, 0, 0, 0, &time.Location{}),
			Platforms: map[string]AppLocation{
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

func Test_patchReleaseURL(t *testing.T) {
	testCases := []struct {
		finalURL string
		expect   autogold.Value
	}{
		{
			finalURL: "https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.21%2B1321.8c3a4999f2/Cody.2023.6.21%2B1321.8c3a4999f2.aarch64.app.tar.gz",
			expect:   autogold.Expect("https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.21%2B1321.8c3a4999f2/Cody_2023.6.21%2B1321.8c3a4999f2_aarch64.dmg"),
		},
		{
			finalURL: "https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.21%2B1321.8c3a4999f2/Cody.2023.6.21%2B1321.8c3a4999f2.x86_64.app.tar.gz",
			expect:   autogold.Expect("https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.21%2B1321.8c3a4999f2/Cody_2023.6.21%2B1321.8c3a4999f2_x64.dmg"),
		},
		{
			finalURL: "https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.21%2B1321.8c3a4999f2/cody_2023.6.21%2B1321.8c3a4999f2_amd64.AppImage.tar.gz",
			expect:   autogold.Expect("https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.21%2B1321.8c3a4999f2/cody_2023.6.21%2B1321.8c3a4999f2_amd64.AppImage.tar.gz"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.finalURL, func(t *testing.T) {
			got := patchReleaseURL(tc.finalURL)
			tc.expect.Equal(t, got)
		})
	}
}
