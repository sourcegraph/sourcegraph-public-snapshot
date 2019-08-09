package registry

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestGetExtensionManifestWithBundleURL(t *testing.T) {
	resetMocks()
	ctx := context.Background()

	nilOrEmpty := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	t0 := time.Unix(1234, 0)

	t.Run(`manifest with "url"`, func(t *testing.T) {
		mocks.releases.GetLatest = func(registryExtensionID int32, releaseTag string, includeArtifacts bool) (*dbRelease, error) {
			return &dbRelease{
				Manifest:  `{"name":"x","url":"u"}`,
				CreatedAt: t0,
			}, nil
		}
		defer func() { mocks.releases.GetLatest = nil }()
		manifest, publishedAt, err := getExtensionManifestWithBundleURL(ctx, "x", 1, "t")
		if err != nil {
			t.Fatal(err)
		}
		if want := `{"name":"x","url":"u"}`; manifest == nil || !jsonDeepEqual(*manifest, want) {
			t.Errorf("got %q, want %q", nilOrEmpty(manifest), want)
		}
		if publishedAt != t0 {
			t.Errorf("got %v, want %v", publishedAt, t0)
		}
	})

	t.Run(`manifest without "url"`, func(t *testing.T) {
		mocks.releases.GetLatest = func(registryExtensionID int32, releaseTag string, includeArtifacts bool) (*dbRelease, error) {
			return &dbRelease{
				Manifest:  `{"name":"x"}`,
				CreatedAt: t0,
			}, nil
		}
		defer func() { mocks.releases.GetLatest = nil }()
		manifest, publishedAt, err := getExtensionManifestWithBundleURL(ctx, "x", 1, "t")
		if err != nil {
			t.Fatal(err)
		}
		if want := `{"name":"x","url":"/-/static/extension/0-x.js?fqw3qlts--x"}`; manifest == nil || !jsonDeepEqual(*manifest, want) {
			t.Errorf("got %q, want %q", nilOrEmpty(manifest), want)
		}
		if publishedAt != t0 {
			t.Errorf("got %v, want %v", publishedAt, t0)
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_682(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
