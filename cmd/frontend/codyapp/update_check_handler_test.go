package codyapp

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
)

var integrationTest = flag.Bool("IntegrationTest", false, "access external services like GCP")

func TestAppVersionPlatformFormat(t *testing.T) {
	tt := []struct {
		Arch   string
		Target string
		Wanted string
	}{
		{
			Arch:   "x86_64",
			Target: "linux",
			Wanted: "x86_64-linux",
		},
		{
			Arch:   "x86_64",
			Target: "darwin",
			Wanted: "x86_64-darwin",
		},
		{
			Arch:   "aarch64",
			Target: "darwin",
			Wanted: "aarch64-darwin",
		},
	}

	for _, tc := range tt {
		appVersion := AppVersion{
			Target:  tc.Target,
			Version: "0.0.0+dev",
			Arch:    tc.Arch,
		}

		if appVersion.Platform() != tc.Wanted {
			t.Errorf("incorrect plaform format - got %q wanted %q", appVersion.Platform(), tc.Wanted)
		}
	}
}

func TestReadAppClientVersion(t *testing.T) {
	var tt = []struct {
		Name    string
		Valid   bool
		Target  string
		Arch    string
		Version string
	}{
		{
			Name:    "client versions gets created from query params",
			Valid:   true,
			Target:  "Darwin",
			Arch:    "x86_64-amd64",
			Version: "1.8.9+debug",
		},
		{
			Name:    "empty target is invalid",
			Valid:   false,
			Target:  "",
			Arch:    "x86_64-amd64",
			Version: "1.8.9+insiders.FFAA",
		},
		{
			Name:    "empty arch is invalid",
			Valid:   false,
			Target:  "Toaster",
			Arch:    "",
			Version: "1.8.9+1234.cc11bbaa",
		},
		{
			Name:    "empty version is invalid",
			Valid:   false,
			Target:  "Kettle",
			Arch:    "x86_64-amd64",
			Version: "",
		},
	}
	reqURL, err := url.Parse("/app/check/update")
	if err != nil {
		t.Fatal("failed to create app update url", err)
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var v = url.Values{}
			v.Add("target", tc.Target)
			v.Add("arch", tc.Arch)

			// we concat the version here since Tauri does not URL encode the version correctly
			reqURL.RawQuery = v.Encode() + "&current_version=" + tc.Version

			appVersion := readClientAppVersion(reqURL)
			validationErr := appVersion.validate()
			if tc.Valid && validationErr != nil {
				t.Errorf("app version failed validation and should have passed - err=%s, appVersion=%v", err, appVersion)
			} else if !tc.Valid && validationErr == nil {
				t.Errorf("invalid app version passed validation - err=%s, appVersion=%v", err, appVersion)
			}
		})
	}
}

func TestAppUpdateCheckHandler(t *testing.T) {
	var resolver = StaticManifestResolver{
		Manifest: AppUpdateManifest{
			Version: "3023.5.8", // set the year part of the version FAR ahead so that there is always a version to update to
			Notes:   "This is a test",
			PubDate: time.Date(2023, time.May, 8, 12, 0, 0, 0, &time.Location{}),
			Platforms: map[string]AppLocation{
				"x86_64-unknown-linux-gnu": {
					Signature: "Yippy Kay YAY",
					URL:       "https://example.com",
				},
			},
		},
	}

	t.Run("with static manifest resolver, and exact version", func(t *testing.T) {
		req, err := clientVersionRequest(t, "unknown-linux-gnu", "x86_64", resolver.Manifest.Version+"+1234.DEADBEEF")
		if err != nil {
			t.Fatalf("failed to create client version request: %v", err)
		}
		w := httptest.NewRecorder()

		checker := NewAppUpdateChecker(logtest.NoOp(t), &resolver)
		checker.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected HTTP Status %d for exact version match, but got %d", http.StatusNoContent, resp.StatusCode)
		}
	})
	t.Run("with static manifest resolver, and older version", func(t *testing.T) {
		var clientVersion = AppVersion{
			Target: "unknown-linux-gnu",
			// this version has to be higher than 2023.6.13 since versions before that are not allowed to update!
			Version: "2023.8.23+old.1234",
			Arch:    "x86_64",
		}

		req, err := clientVersionRequest(t, clientVersion.Target, clientVersion.Arch, clientVersion.Version)
		if err != nil {
			t.Fatalf("failed to create client version request: %v", err)
		}

		w := httptest.NewRecorder()

		checker := NewAppUpdateChecker(logtest.Scoped(t), &resolver)
		checker.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected HTTP Status %d for exact version match, but got %d", http.StatusOK, resp.StatusCode)
		}

		var updateResp AppUpdateResponse
		err = json.NewDecoder(resp.Body).Decode(&updateResp)
		if err != nil {
			t.Fatalf("failed to decode AppUpdateManifest: %v", err)
		}

		if resolver.Manifest.Version != updateResp.Version {
			t.Errorf("Wanted %s manifest version, got %s", resolver.Manifest.Version, updateResp.Version)
		}
		if resolver.Manifest.PubDate.String() != updateResp.PubDate.String() {
			t.Errorf("Wanted %s manifest version, got %s", resolver.Manifest.Version, updateResp.Version)
		}

		if platform, ok := resolver.Manifest.Platforms[clientVersion.Platform()]; !ok {
			t.Fatalf("failed to get %q platform from manifest", clientVersion.Platform())
		} else if updateResp.Signature != platform.Signature {
			t.Errorf("signature mismatch. Got %q wanted %q", updateResp.Signature, platform.Signature)
		} else if updateResp.URL != platform.URL {
			t.Errorf("URL mismatch. Got %q wanted %q", updateResp.URL, platform.URL)
		}
	})
	t.Run("client on or before '2023.6.13' gets told there are no updates", func(t *testing.T) {
		noUpdateVersions := []string{"2023.6.13+1234.stuff", "2021.1.11+1234.stuff"}

		for _, version := range noUpdateVersions {
			req, err := clientVersionRequest(t, "unknown-linux-gnu", "x86_64", version)
			if err != nil {
				t.Fatalf("failed to create client version request: %v", err)
			}
			w := httptest.NewRecorder()

			checker := NewAppUpdateChecker(logtest.NoOp(t), &resolver)
			checker.Handler().ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusNoContent {
				t.Errorf("expected HTTP Status %d for client on version %s (who should not receive updates) but got %d", http.StatusNoContent, version, resp.StatusCode)
			}
		}
	})
}

func TestGCSResolver(t *testing.T) {
	flag.Parse()

	if !*integrationTest {
		t.Skip("integration testing is not enabled - to enable this test pass the flag '-IntegrationTest'")
		return
	}

	ctx := context.Background()
	resolver, err := NewGCSManifestResolver(ctx, ManifestBucket, ManifestName)
	if err != nil {
		t.Fatalf("failed to create GCS manifest resolver: %v", err)
	}

	gcsManifest, err := resolver.Resolve(ctx)
	if err != nil {
		t.Fatalf("failed to get manifest using GCS resolver: %v", err)
	}

	if gcsManifest == nil {
		t.Errorf("got nil Version Manifest")
	}

	if gcsManifest.Version == "" {
		t.Errorf("GCS Manifest Version is empty")
	}
	if gcsManifest.PubDate.IsZero() {
		t.Errorf("GCS Manifest PubDate is Zero: %s", gcsManifest.PubDate.String())
	}

	if len(gcsManifest.Platforms) == 0 {
		t.Errorf("GCS Manifest has zero platforms: %v", gcsManifest)
	}

	for keyPlatform, got := range gcsManifest.Platforms {
		if got.Signature == "" {
			t.Errorf("%s platform has an empty signature", keyPlatform)
		}
		if got.URL == "" {
			t.Errorf("%s platform has an empty url", keyPlatform)
		}
	}

}

func clientVersionRequest(t *testing.T, target, arch, version string) (*http.Request, error) {
	t.Helper()
	var v = url.Values{}
	v.Add("target", target)
	v.Add("arch", arch)
	reqURL, err := url.Parse("http://localhost")
	if err != nil {
		return nil, err
	}
	// we concat the version here since Tauri does not URL encode the version correctly
	reqURL.RawQuery = v.Encode() + "&current_version=" + version
	return httptest.NewRequest("GET", reqURL.String(), nil), nil
}
