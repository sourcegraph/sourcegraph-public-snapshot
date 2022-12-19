package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestServiceConnections(t *testing.T) {
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// We override the URLs so service discovery doesn't try and talk to k8s
	oldSearcherURL := searcherURL
	t.Cleanup(func() { searcherURL = oldSearcherURL })
	searcherURL = "http://searcher:3181"

	indexedKey := "INDEXED_SEARCH_SERVERS"
	oldIndexedSearchServers := os.Getenv(indexedKey)
	t.Cleanup(func() { os.Setenv(indexedKey, oldIndexedSearchServers) })
	os.Setenv(indexedKey, "http://indexed-search:6070")

	// We only test that we get something non-empty back.
	sc := serviceConnections(logtest.Scoped(t))
	if reflect.DeepEqual(sc, conftypes.ServiceConnections{}) {
		t.Fatal("expected non-empty service connections")
	}
}

func TestServiceConnectionsZoektsIntentionallyEmpty(t *testing.T) {
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// We override the URLs so service discovery doesn't try and talk to k8s
	oldSearcherURL := searcherURL
	t.Cleanup(func() { searcherURL = oldSearcherURL })
	searcherURL = "http://searcher:3181"

	indexedKey := "INDEXED_SEARCH_SERVERS"
	oldIndexedSearchServers := os.Getenv(indexedKey)
	t.Cleanup(func() { os.Setenv(indexedKey, oldIndexedSearchServers) })
	os.Setenv(indexedKey, "")

	sc := serviceConnections(logtest.Scoped(t))
	if sc.Zoekts != nil {
		t.Errorf("Expected zero Zoekt service connections but identified %v", len(sc.Zoekts))
	}
	if !sc.ZoektsIntentionallyEmpty {
		t.Error("Expected ZoektsIntentionallyEmpty")
	}
}

func TestWriteSiteConfig(t *testing.T) {
	db := database.NewMockDB()
	confStore := database.NewMockConfStore()
	conf := &database.SiteConfig{ID: 1}
	confStore.SiteGetLatestFunc.SetDefaultReturn(
		conf,
		nil,
	)
	logger := logtest.Scoped(t)
	db.ConfFunc.SetDefaultReturn(confStore)
	confSource := newConfigurationSource(logger, db)

	t.Run("error when incorrect last ID", func(t *testing.T) {
		err := confSource.Write(context.Background(), conftypes.RawUnified{}, conf.ID-1)
		assert.Error(t, err)
	})

	t.Run("no error when correct last ID", func(t *testing.T) {
		err := confSource.Write(context.Background(), conftypes.RawUnified{}, conf.ID)
		assert.NoError(t, err)
	})
}

func TestReadSiteConfigFile(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		Name  string
		Files []string
		Want  string
		Err   string
	}{{
		Name:  "one",
		Files: []string{`{"hello": "world"}`},
		Want:  `{"hello": "world"}`,
	}, {
		Name: "two",
		Files: []string{
			`// leading comment
{
  // first comment
  "first": "file",
} // trailing comment
`, `{"second": "file"}`},
		Want: `// merged SITE_CONFIG_FILE
{
  // BEGIN $tmp/0.json
  "first": "file",
  // END $tmp/0.json
  // BEGIN $tmp/1.json
  "second": "file",
  // END $tmp/1.json
}`,
	},
		{
			Name: "three",
			Files: []string{
				`{
    "search.index.branches": {
      "github.com/sourcegraph/sourcegraph": ["3.17", "v3.0.0"],
      "github.com/kubernetes/kubernetes": ["release-1.17"],
      "github.com/go-yaml/yaml": ["v2", "v3"]
    }
}`,
				`{
  "observability.alerts": [ {"level":"warning"}, { "level": "critical"} ]
}`},
			Want: `// merged SITE_CONFIG_FILE
{
  // BEGIN $tmp/0.json
  "search.index.branches": {
    "github.com/go-yaml/yaml": [
      "v2",
      "v3"
    ],
    "github.com/kubernetes/kubernetes": [
      "release-1.17"
    ],
    "github.com/sourcegraph/sourcegraph": [
      "3.17",
      "v3.0.0"
    ]
  },
  // END $tmp/0.json
  // BEGIN $tmp/1.json
  "observability.alerts": [
    {
      "level": "warning"
    },
    {
      "level": "critical"
    }
  ],
  // END $tmp/1.json
}`,
		},
		{
			Name: "parse-error",
			Files: []string{
				"{}",
				"{",
			},
			Err: "CloseBraceExpected",
		}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var paths []string
			for i, b := range c.Files {
				p := filepath.Join(dir, fmt.Sprintf("%d.json", i))
				paths = append(paths, p)
				if err := os.WriteFile(p, []byte(b), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := readSiteConfigFile(paths)
			if c.Err != "" && !strings.Contains(fmt.Sprintf("%s", err), c.Err) {
				t.Fatalf("%s doesn't contain error substring %s", err, c.Err)
			}
			got = bytes.ReplaceAll(got, []byte(dir), []byte("$tmp"))
			if d := cmp.Diff(c.Want, string(got)); d != "" {
				t.Fatalf("unexpected merge (-want, +got):\n%s", d)
			}
		})
	}
}

func TestZoektAddr(t *testing.T) {
	cases := []struct {
		name    string
		environ []string
		want    string
	}{{
		name: "default",
		want: "k8s+rpc://indexed-search:6070?kind=sts",
	}, {
		name:    "old",
		environ: []string{"ZOEKT_HOST=127.0.0.1:3070"},
		want:    "127.0.0.1:3070",
	}, {
		name:    "new",
		environ: []string{"INDEXED_SEARCH_SERVERS=indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070"},
		want:    "indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
	}, {
		name: "prefer new",
		environ: []string{
			"ZOEKT_HOST=127.0.0.1:3070",
			"INDEXED_SEARCH_SERVERS=indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
		},
		want: "indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
	}, {
		name: "unset new",
		environ: []string{
			"ZOEKT_HOST=127.0.0.1:3070",
			"INDEXED_SEARCH_SERVERS=",
		},
		want: "",
	}, {
		name: "unset old",
		environ: []string{
			"ZOEKT_HOST=",
		},
		want: "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := zoektAddr(tc.environ)
			if got != tc.want {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
