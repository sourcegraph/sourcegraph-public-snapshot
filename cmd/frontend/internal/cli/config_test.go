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
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestServiceConnections(t *testing.T) {
	t.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// We override the URLs so service discovery doesn't try and talk to k8s
	searcherKey := "SEARCHER_URL"
	t.Setenv(searcherKey, "http://searcher:3181")

	indexedKey := "INDEXED_SEARCH_SERVERS"
	t.Setenv(indexedKey, "http://indexed-search:6070")

	// We only test that we get something non-empty back.
	sc := serviceConnections(logtest.Scoped(t))
	if reflect.DeepEqual(sc, conftypes.ServiceConnections{}) {
		t.Fatal("expected non-empty service connections")
	}
}

func TestWriteSiteConfig(t *testing.T) {
	db := dbmocks.NewMockDB()
	confStore := dbmocks.NewMockConfStore()
	conf := &database.SiteConfig{ID: 1}
	confStore.SiteGetLatestFunc.SetDefaultReturn(
		conf,
		nil,
	)
	logger := logtest.Scoped(t)
	db.ConfFunc.SetDefaultReturn(confStore)
	confSource := newConfigurationSource(logger, db)

	t.Run("error when incorrect last ID", func(t *testing.T) {
		err := confSource.Write(context.Background(), conftypes.RawUnified{}, conf.ID-1, 0)
		assert.Error(t, err)
	})

	t.Run("no error when correct last ID", func(t *testing.T) {
		err := confSource.Write(context.Background(), conftypes.RawUnified{}, conf.ID, 1)
		assert.NoError(t, err)
	})
}

func TestOverrideSiteConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	conf := db.Conf()
	ctx := context.Background()

	// Required so that the serviceConnections function when called won't panic.
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// Generate a SITE_CONFIG_FILE and set the env.
	file, err := os.CreateTemp("", "site-config-*")
	require.NoError(t, err, "failed to create temp file")

	_, err = file.Write([]byte(`{"auth.providers": [{ "type": "builtin"}]}`))
	require.NoError(t, err, "failed to write to temp file")

	os.Setenv("SITE_CONFIG_FILE", file.Name())

	// First time a write happens.
	err = overrideSiteConfig(ctx, logger, db)
	require.NoError(t, err, "failed to override ")

	// Read the latest site config.
	current, err := conf.SiteGetLatest(ctx)
	require.NoError(t, err, "failed to read site config")

	// Try to write again which would happen if Sourcegraph is restarted for example.
	err = overrideSiteConfig(ctx, logger, db)
	require.NoError(t, err, "failed to override ")

	next, err := conf.SiteGetLatest(ctx)
	require.NoError(t, err, "failed to read site config")

	// Since the SITE_CONFIG_FILE has not changed, expect no changes.
	require.Equal(t, current.ID, next.ID)

	// Now create a new SITE_CONFIG_FILE with a different config and update the env variable.
	file2, err := os.CreateTemp("", "site-config-*")
	require.NoError(t, err, "failed to create temp file")

	_, err = file2.Write([]byte(`{"auth.providers": [{"type": "builtin"}], "disableAutoGitUpdates": true}`))
	require.NoError(t, err, "failed to write to temp file")

	os.Setenv("SITE_CONFIG_FILE", file2.Name())

	// Try to write again.
	err = overrideSiteConfig(ctx, logger, db)
	require.NoError(t, err, "failed to override ")

	// Read the latest site config.
	next2, err := conf.SiteGetLatest(ctx)
	require.NoError(t, err, "failed to read site config")

	// Since the SITE_CONFIG_FILE has changed, expect a new ID.
	require.NotEqual(t, next.ID, next2.ID)
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

func TestGitserverAddr(t *testing.T) {
	cases := []struct {
		name       string
		deployType string
		environ    []string
		want       string
		wantErr    error
	}{{
		name: "test default",
		want: "gitserver:3178",
	}, {
		name:    "default",
		environ: []string{"SRC_GIT_SERVERS=k8s+rpc://gitserver:3178?kind=sts"},
		want:    "k8s+rpc://gitserver:3178?kind=sts",
	}, {
		name:    "exact",
		environ: []string{"SRC_GIT_SERVERS=gitserver-0:3178 gitserver-1:3178"},
		want:    "gitserver-0:3178 gitserver-1:3178",
	}, {
		name: "replicas",
		environ: []string{
			"SRC_GIT_SERVERS=2",
		},
		want: "gitserver-0.gitserver:3178 gitserver-1.gitserver:3178",
	}, {
		name:       "replicas helm",
		deployType: deploy.Helm,
		environ: []string{
			"SRC_GIT_SERVERS=2",
		},
		want: "gitserver-0.gitserver:3178 gitserver-1.gitserver:3178",
	}, {
		name:       "replicas docker-compose",
		deployType: deploy.DockerCompose,
		environ: []string{
			"SRC_GIT_SERVERS=2",
		},
		want: "gitserver-0:3178 gitserver-1:3178",
	}, {
		name:       "unsupported deploy type",
		deployType: deploy.PureDocker,
		environ: []string{
			"SRC_GIT_SERVERS=5",
		},
		wantErr: errors.New("unsupported deployment type: pure-docker"),
	}, {
		name: "unset",
		environ: []string{
			"SRC_GIT_SERVERS=",
		},
		want: "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := gitserverAddr(tc.environ)
			if (err != nil) != (tc.wantErr != nil) {
				t.Fatalf("got err %v, want %s", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestSearcherAddr(t *testing.T) {
	cases := []struct {
		name       string
		deployType string
		environ    []string
		want       string
		wantErr    error
	}{{
		name: "default",
		want: "k8s+http://searcher:3181",
	}, {
		name:    "stateful",
		environ: []string{"SEARCHER_URL=k8s+rpc://searcher:3181?kind=sts"},
		want:    "k8s+rpc://searcher:3181?kind=sts",
	}, {
		name:    "exact",
		environ: []string{"SEARCHER_URL=http://searcher-0:3181 http://searcher-1:3181"},
		want:    "http://searcher-0:3181 http://searcher-1:3181",
	}, {
		name:    "replicas",
		environ: []string{"SEARCHER_URL=2"},
		want:    "http://searcher-0.searcher:3181 http://searcher-1.searcher:3181",
	}, {
		name:       "replicas helm",
		deployType: deploy.Helm,
		environ:    []string{"SEARCHER_URL=2"},
		want:       "http://searcher-0.searcher:3181 http://searcher-1.searcher:3181",
	}, {
		name:       "replicas docker-compose",
		deployType: deploy.DockerCompose,
		environ:    []string{"SEARCHER_URL=2"},
		want:       "http://searcher-0:3181 http://searcher-1:3181",
	}, {
		name:       "unsupported deploy type",
		deployType: deploy.PureDocker,
		environ:    []string{"SEARCHER_URL=5"},
		wantErr:    errors.New("unsupported deployment type: pure-docker"),
	}, {
		name:    "unset",
		environ: []string{"SEARCHER_URL="},
		want:    "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := searcherAddr(tc.environ)
			if (err != nil) != (tc.wantErr != nil) {
				t.Fatalf("got err %v, want %s", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestSymbolsAddr(t *testing.T) {
	cases := []struct {
		name       string
		deployType string
		environ    []string
		want       string
		wantErr    error
	}{{
		name: "default",
		want: "http://symbols:3184",
	}, {
		name:    "stateful",
		environ: []string{"SYMBOLS_URL=k8s+rpc://symbols:3184?kind=sts"},
		want:    "k8s+rpc://symbols:3184?kind=sts",
	}, {
		name:    "exact",
		environ: []string{"SYMBOLS_URL=http://symbols-0:3184 http://symbols-1:3184"},
		want:    "http://symbols-0:3184 http://symbols-1:3184",
	}, {
		name:    "replicas",
		environ: []string{"SYMBOLS_URL=2"},
		want:    "http://symbols-0.symbols:3184 http://symbols-1.symbols:3184",
	}, {
		name:       "replicas kustomize",
		deployType: deploy.Kustomize,
		environ:    []string{"SYMBOLS_URL=2"},
		want:       "http://symbols-0.symbols:3184 http://symbols-1.symbols:3184",
	}, {
		name:       "replicas docker-compose",
		deployType: deploy.DockerCompose,
		environ:    []string{"SYMBOLS_URL=2"},
		want:       "http://symbols-0:3184 http://symbols-1:3184",
	}, {
		name:       "ignore duplicate",
		deployType: deploy.DockerCompose,
		environ:    []string{"SYMBOLS_URL=k8s+rpc://symbols:3184?kind=sts", "SYMBOLS_URL=2"},
		want:       "k8s+rpc://symbols:3184?kind=sts",
	}, {
		name:       "unsupported deploy type",
		deployType: deploy.SingleDocker,
		environ:    []string{"SYMBOLS_URL=2"},
		wantErr:    errors.New("unsupported deployment type: single-docker"),
	}, {
		name:    "unset",
		environ: []string{"SYMBOLS_URL="},
		want:    "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := symbolsAddr(tc.environ)
			if (err != nil) != (tc.wantErr != nil) {
				t.Fatalf("got err %v, want %s", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestZoektAddr(t *testing.T) {
	cases := []struct {
		name       string
		deployType string
		environ    []string
		want       string
		wantErr    error
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
		name:    "replicas",
		environ: []string{"INDEXED_SEARCH_SERVERS=2"},
		want:    "indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
	}, {
		name:       "replicas helm",
		deployType: deploy.Helm,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2"},
		want:       "indexed-search-0.indexed-search:6070 indexed-search-1.indexed-search:6070",
	}, {
		name:       "replicas docker-compose",
		deployType: deploy.DockerCompose,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2"},
		want:       "zoekt-webserver-0:6070 zoekt-webserver-1:6070",
	}, {
		name:    "unset new",
		environ: []string{"ZOEKT_HOST=127.0.0.1:3070", "INDEXED_SEARCH_SERVERS="},
		want:    "",
	}, {
		name:       "ignore duplicate",
		deployType: deploy.DockerCompose,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2", "INDEXED_SEARCH_SERVERS=k8s+rpc://indexed-search:6070?kind=sts"},
		want:       "zoekt-webserver-0:6070 zoekt-webserver-1:6070",
	}, {
		name:       "unsupported deploy type",
		deployType: deploy.SingleDocker,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2"},
		wantErr:    errors.New("unsupported deployment type: single-docker"),
	}, {
		name:    "unset old",
		environ: []string{"ZOEKT_HOST="},
		want:    "",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := zoektAddr(tc.environ)
			if (err != nil) != (tc.wantErr != nil) {
				t.Fatalf("got err %v, want %s", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
