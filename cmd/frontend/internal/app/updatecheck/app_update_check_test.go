package updatecheck

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
			Version: "1.8.9",
		},
		{
			Name:    "empty target is invalid",
			Valid:   false,
			Target:  "",
			Arch:    "x86_64-amd64",
			Version: "1.8.9",
		},
		{
			Name:    "empty arch is invalid",
			Valid:   false,
			Target:  "Toaster",
			Arch:    "",
			Version: "1.8.9",
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
			v.Add("current_version", tc.Version)

			reqURL.RawQuery = v.Encode()

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
		manifest: AppUpdateManifest{
			Version: "2023.5.8",
			Notes:   "This is a test",
			PubDate: time.Date(2023, time.May, 8, 12, 0, 0, 0, &time.Location{}),
			Platforms: map[string]AppLocation{
				"x86_64-amd64": {
					Signature: "Yippy Kay YAY",
					Url:       "https://example.com",
				},
			},
		},
	}

	t.Run("with static manifest resolver, and exact version", func(t *testing.T) {
		var v = url.Values{}
		v.Add("target", "123")
		v.Add("arch", "x86_64-amd64")
		v.Add("current_version", "2023.5.8")
		reqURL, err := url.Parse("http://localhost")
		if err != nil {
			t.Fatalf("failed to parse test server url: %v", err)
		}
		reqURL.RawQuery = v.Encode()
		req := httptest.NewRequest("GET", reqURL.String(), nil)
		w := httptest.NewRecorder()

		checker := NewAppUpdateChecker(logtest.NoOp(t), &resolver)
		checker.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected HTTP Status %d for exact version match, but got %d", http.StatusNoContent, resp.StatusCode)
		}
	})
	t.Run("with static manifest resolver, and older version", func(t *testing.T) {
		var v = url.Values{}
		v.Add("target", "123")
		v.Add("arch", "x86_64-amd64")
		v.Add("current_version", "2000.3.4")
		reqURL, err := url.Parse("http://localhost")
		if err != nil {
			t.Fatalf("failed to parse test server url: %v", err)
		}
		reqURL.RawQuery = v.Encode()
		req := httptest.NewRequest("GET", reqURL.String(), nil)
		w := httptest.NewRecorder()

		checker := NewAppUpdateChecker(logtest.NoOp(t), &resolver)
		checker.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected HTTP Status %d for exact version match, but got %d", http.StatusNoContent, resp.StatusCode)
		}

		var manifest AppUpdateManifest
		err = json.NewDecoder(resp.Body).Decode(&manifest)
		if err != nil {
			t.Fatalf("failed to decode AppUpdateManifest: %v", err)
		}

		if resolver.manifest.Version != manifest.Version {
			t.Errorf("Wanted %s manifest version, got %s", resolver.manifest.Version, manifest.Version)
		}
		if resolver.manifest.PubDate.String() != manifest.PubDate.String() {
			t.Errorf("Wanted %s manifest version, got %s", resolver.manifest.Version, manifest.Version)
		}

		for k, wanted := range resolver.manifest.Platforms {
			got := manifest.Platforms[k]
			if wanted.Signature != got.Signature {
				t.Errorf("signature mistmactch - got %s wanted %s", got.Signature, wanted.Signature)
			}
			if wanted.Url != got.Url {
				t.Errorf("url mistmactch - got %s wanted %s", got.Signature, wanted.Signature)
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
		if got.Url == "" {
			t.Errorf("%s platform has an empty url", keyPlatform)
		}
	}

}
