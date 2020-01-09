package srccli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Masterminds/semver"
)

func TestHighestMatchingVersion(t *testing.T) {
	minimum := semver.MustParse("1.2.3")
	versions := []*semver.Version{
		semver.MustParse("1.1.4"),
		semver.MustParse("1.2.2"),
		semver.MustParse("1.2.3"),
		semver.MustParse("1.2.5"),
		semver.MustParse("1.3.1"),
	}

	version, err := highestMatchingVersion(minimum, versions)
	if err != nil {
		t.Fatal(err)
	}

	if version.String() != "1.2.5" {
		t.Errorf("got %s, want 1.2.5", version.String())
	}
}

func TestReleaseVersions(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "" {
			// Page 1
			w.Header().Add("Link", fmt.Sprintf(`<%s?page=2>; rel="next"`, server.URL))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"tag_name": "1.1.1"},
				{"tag_name": "1.1.2"},
				{"tag_name": "1.1.3"},
				{"tag_name": "1.1.4", "draft": true},
				{"tag_name": "1.1.5", "prerelease": true}
			]`))
		} else {
			// Page 2
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"tag_name": "latest"},
				{"tag_name": "1.2.2", "draft": true},
				{"tag_name": "1.2.3"},
				{"tag_name": "1.2.4", "prerelease": true},
				{"tag_name": "1.2.5"}
			]`))
		}
	}))
	defer server.Close()

	versions, err := releaseVersions(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	expectedVersions := map[string]bool{
		"1.1.1": true,
		"1.1.2": true,
		"1.1.3": true,
		"1.2.3": true,
		"1.2.5": true,
	}

	if len(versions) != 5 {
		t.Errorf("got %d versions, want 5", len(versions))
	}

	for _, v := range versions {
		if _, ok := expectedVersions[v.String()]; !ok {
			t.Errorf("unexpected version %s", v.String())
		}
	}
}
