package backend

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func TestGetExtensionManifestWithBundleURL(t *testing.T) {
	ctx := testContext()

	nilOrEmpty := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	t.Run(`manifest with "url"`, func(t *testing.T) {
		db.Mocks.RegistryExtensionReleases.GetLatest = func(registryExtensionID int32, releaseTag string, includeArtifacts bool) (*db.RegistryExtensionRelease, error) {
			return &db.RegistryExtensionRelease{
				Manifest: `{"name":"x","url":"u"}`,
			}, nil
		}
		defer func() { db.Mocks.RegistryExtensionReleases.GetLatest = nil }()
		manifest, err := GetExtensionManifestWithBundleURL(ctx, "x", 1, "t")
		if err != nil {
			t.Fatal(err)
		}
		if want := `{"name":"x","url":"u"}`; manifest == nil || !jsonDeepEqual(*manifest, want) {
			t.Errorf("got %q, want %q", nilOrEmpty(manifest), want)
		}
	})

	t.Run(`manifest without "url"`, func(t *testing.T) {
		db.Mocks.RegistryExtensionReleases.GetLatest = func(registryExtensionID int32, releaseTag string, includeArtifacts bool) (*db.RegistryExtensionRelease, error) {
			return &db.RegistryExtensionRelease{
				Manifest: `{"name":"x"}`,
			}, nil
		}
		defer func() { db.Mocks.RegistryExtensionReleases.GetLatest = nil }()
		manifest, err := GetExtensionManifestWithBundleURL(ctx, "x", 1, "t")
		if err != nil {
			t.Fatal(err)
		}
		if want := `{"name":"x","url":"/-/static/extension/0.js?-1fmlvpbbdw2yo#x"}`; manifest == nil || !jsonDeepEqual(*manifest, want) {
			t.Errorf("got %q, want %q", nilOrEmpty(manifest), want)
		}
	})
}

func jsonDeepEqual(a, b string) bool {
	var va, vb interface{}
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		panic(err)
	}
	return reflect.DeepEqual(va, vb)
}
