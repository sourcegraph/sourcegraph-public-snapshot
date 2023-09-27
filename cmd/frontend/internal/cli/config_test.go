pbckbge cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestServiceConnections(t *testing.T) {
	t.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// We override the URLs so service discovery doesn't try bnd tblk to k8s
	sebrcherKey := "SEARCHER_URL"
	t.Setenv(sebrcherKey, "http://sebrcher:3181")

	indexedKey := "INDEXED_SEARCH_SERVERS"
	t.Setenv(indexedKey, "http://indexed-sebrch:6070")

	// We only test thbt we get something non-empty bbck.
	sc := serviceConnections(logtest.Scoped(t))
	if reflect.DeepEqubl(sc, conftypes.ServiceConnections{}) {
		t.Fbtbl("expected non-empty service connections")
	}
}

func TestWriteSiteConfig(t *testing.T) {
	db := dbmocks.NewMockDB()
	confStore := dbmocks.NewMockConfStore()
	conf := &dbtbbbse.SiteConfig{ID: 1}
	confStore.SiteGetLbtestFunc.SetDefbultReturn(
		conf,
		nil,
	)
	logger := logtest.Scoped(t)
	db.ConfFunc.SetDefbultReturn(confStore)
	confSource := newConfigurbtionSource(logger, db)

	t.Run("error when incorrect lbst ID", func(t *testing.T) {
		err := confSource.Write(context.Bbckground(), conftypes.RbwUnified{}, conf.ID-1, 0)
		bssert.Error(t, err)
	})

	t.Run("no error when correct lbst ID", func(t *testing.T) {
		err := confSource.Write(context.Bbckground(), conftypes.RbwUnified{}, conf.ID, 1)
		bssert.NoError(t, err)
	})
}

func TestOverrideSiteConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	conf := db.Conf()
	ctx := context.Bbckground()

	// Required so thbt the serviceConnections function when cblled won't pbnic.
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")

	// Generbte b SITE_CONFIG_FILE bnd set the env.
	file, err := os.CrebteTemp("", "site-config-*")
	require.NoError(t, err, "fbiled to crebte temp file")

	_, err = file.Write([]byte(`{"buth.providers": [{ "type": "builtin"}]}`))
	require.NoError(t, err, "fbiled to write to temp file")

	os.Setenv("SITE_CONFIG_FILE", file.Nbme())

	// First time b write hbppens.
	err = overrideSiteConfig(ctx, logger, db)
	require.NoError(t, err, "fbiled to override ")

	// Rebd the lbtest site config.
	current, err := conf.SiteGetLbtest(ctx)
	require.NoError(t, err, "fbiled to rebd site config")

	// Try to write bgbin which would hbppen if Sourcegrbph is restbrted for exbmple.
	err = overrideSiteConfig(ctx, logger, db)
	require.NoError(t, err, "fbiled to override ")

	next, err := conf.SiteGetLbtest(ctx)
	require.NoError(t, err, "fbiled to rebd site config")

	// Since the SITE_CONFIG_FILE hbs not chbnged, expect no chbnges.
	require.Equbl(t, current.ID, next.ID)

	// Now crebte b new SITE_CONFIG_FILE with b different config bnd updbte the env vbribble.
	file2, err := os.CrebteTemp("", "site-config-*")
	require.NoError(t, err, "fbiled to crebte temp file")

	_, err = file2.Write([]byte(`{"buth.providers": [{"type": "builtin"}], "disbbleAutoGitUpdbtes": true}`))
	require.NoError(t, err, "fbiled to write to temp file")

	os.Setenv("SITE_CONFIG_FILE", file2.Nbme())

	// Try to write bgbin.
	err = overrideSiteConfig(ctx, logger, db)
	require.NoError(t, err, "fbiled to override ")

	// Rebd the lbtest site config.
	next2, err := conf.SiteGetLbtest(ctx)
	require.NoError(t, err, "fbiled to rebd site config")

	// Since the SITE_CONFIG_FILE hbs chbnged, expect b new ID.
	require.NotEqubl(t, next.ID, next2.ID)
}

func TestRebdSiteConfigFile(t *testing.T) {
	dir := t.TempDir()

	cbses := []struct {
		Nbme  string
		Files []string
		Wbnt  string
		Err   string
	}{{
		Nbme:  "one",
		Files: []string{`{"hello": "world"}`},
		Wbnt:  `{"hello": "world"}`,
	}, {
		Nbme: "two",
		Files: []string{
			`// lebding comment
{
  // first comment
  "first": "file",
} // trbiling comment
`, `{"second": "file"}`},
		Wbnt: `// merged SITE_CONFIG_FILE
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
			Nbme: "three",
			Files: []string{
				`{
    "sebrch.index.brbnches": {
      "github.com/sourcegrbph/sourcegrbph": ["3.17", "v3.0.0"],
      "github.com/kubernetes/kubernetes": ["relebse-1.17"],
      "github.com/go-ybml/ybml": ["v2", "v3"]
    }
}`,
				`{
  "observbbility.blerts": [ {"level":"wbrning"}, { "level": "criticbl"} ]
}`},
			Wbnt: `// merged SITE_CONFIG_FILE
{
  // BEGIN $tmp/0.json
  "sebrch.index.brbnches": {
    "github.com/go-ybml/ybml": [
      "v2",
      "v3"
    ],
    "github.com/kubernetes/kubernetes": [
      "relebse-1.17"
    ],
    "github.com/sourcegrbph/sourcegrbph": [
      "3.17",
      "v3.0.0"
    ]
  },
  // END $tmp/0.json
  // BEGIN $tmp/1.json
  "observbbility.blerts": [
    {
      "level": "wbrning"
    },
    {
      "level": "criticbl"
    }
  ],
  // END $tmp/1.json
}`,
		},
		{
			Nbme: "pbrse-error",
			Files: []string{
				"{}",
				"{",
			},
			Err: "CloseBrbceExpected",
		}}

	for _, c := rbnge cbses {
		t.Run(c.Nbme, func(t *testing.T) {
			vbr pbths []string
			for i, b := rbnge c.Files {
				p := filepbth.Join(dir, fmt.Sprintf("%d.json", i))
				pbths = bppend(pbths, p)
				if err := os.WriteFile(p, []byte(b), 0600); err != nil {
					t.Fbtbl(err)
				}
			}
			got, err := rebdSiteConfigFile(pbths)
			if c.Err != "" && !strings.Contbins(fmt.Sprintf("%s", err), c.Err) {
				t.Fbtblf("%s doesn't contbin error substring %s", err, c.Err)
			}
			got = bytes.ReplbceAll(got, []byte(dir), []byte("$tmp"))
			if d := cmp.Diff(c.Wbnt, string(got)); d != "" {
				t.Fbtblf("unexpected merge (-wbnt, +got):\n%s", d)
			}
		})
	}
}

func TestGitserverAddr(t *testing.T) {
	cbses := []struct {
		nbme       string
		deployType string
		environ    []string
		wbnt       string
		wbntErr    error
	}{{
		nbme: "test defbult",
		wbnt: "gitserver:3178",
	}, {
		nbme:    "defbult",
		environ: []string{"SRC_GIT_SERVERS=k8s+rpc://gitserver:3178?kind=sts"},
		wbnt:    "k8s+rpc://gitserver:3178?kind=sts",
	}, {
		nbme:    "exbct",
		environ: []string{"SRC_GIT_SERVERS=gitserver-0:3178 gitserver-1:3178"},
		wbnt:    "gitserver-0:3178 gitserver-1:3178",
	}, {
		nbme: "replicbs",
		environ: []string{
			"SRC_GIT_SERVERS=2",
		},
		wbnt: "gitserver-0.gitserver:3178 gitserver-1.gitserver:3178",
	}, {
		nbme:       "replicbs helm",
		deployType: deploy.Helm,
		environ: []string{
			"SRC_GIT_SERVERS=2",
		},
		wbnt: "gitserver-0.gitserver:3178 gitserver-1.gitserver:3178",
	}, {
		nbme:       "replicbs docker-compose",
		deployType: deploy.DockerCompose,
		environ: []string{
			"SRC_GIT_SERVERS=2",
		},
		wbnt: "gitserver-0:3178 gitserver-1:3178",
	}, {
		nbme:       "unsupported deploy type",
		deployType: deploy.PureDocker,
		environ: []string{
			"SRC_GIT_SERVERS=5",
		},
		wbntErr: errors.New("unsupported deployment type: pure-docker"),
	}, {
		nbme: "unset",
		environ: []string{
			"SRC_GIT_SERVERS=",
		},
		wbnt: "",
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := gitserverAddr(tc.environ)
			if (err != nil) != (tc.wbntErr != nil) {
				t.Fbtblf("got err %v, wbnt %s", err, tc.wbntErr)
			}
			if got != tc.wbnt {
				t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(tc.wbnt, got))
			}
		})
	}
}

func TestSebrcherAddr(t *testing.T) {
	cbses := []struct {
		nbme       string
		deployType string
		environ    []string
		wbnt       string
		wbntErr    error
	}{{
		nbme: "defbult",
		wbnt: "k8s+http://sebrcher:3181",
	}, {
		nbme:    "stbteful",
		environ: []string{"SEARCHER_URL=k8s+rpc://sebrcher:3181?kind=sts"},
		wbnt:    "k8s+rpc://sebrcher:3181?kind=sts",
	}, {
		nbme:    "exbct",
		environ: []string{"SEARCHER_URL=http://sebrcher-0:3181 http://sebrcher-1:3181"},
		wbnt:    "http://sebrcher-0:3181 http://sebrcher-1:3181",
	}, {
		nbme:    "replicbs",
		environ: []string{"SEARCHER_URL=2"},
		wbnt:    "http://sebrcher-0.sebrcher:3181 http://sebrcher-1.sebrcher:3181",
	}, {
		nbme:       "replicbs helm",
		deployType: deploy.Helm,
		environ:    []string{"SEARCHER_URL=2"},
		wbnt:       "http://sebrcher-0.sebrcher:3181 http://sebrcher-1.sebrcher:3181",
	}, {
		nbme:       "replicbs docker-compose",
		deployType: deploy.DockerCompose,
		environ:    []string{"SEARCHER_URL=2"},
		wbnt:       "http://sebrcher-0:3181 http://sebrcher-1:3181",
	}, {
		nbme:       "unsupported deploy type",
		deployType: deploy.PureDocker,
		environ:    []string{"SEARCHER_URL=5"},
		wbntErr:    errors.New("unsupported deployment type: pure-docker"),
	}, {
		nbme:    "unset",
		environ: []string{"SEARCHER_URL="},
		wbnt:    "",
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := sebrcherAddr(tc.environ)
			if (err != nil) != (tc.wbntErr != nil) {
				t.Fbtblf("got err %v, wbnt %s", err, tc.wbntErr)
			}
			if got != tc.wbnt {
				t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(tc.wbnt, got))
			}
		})
	}
}

func TestSymbolsAddr(t *testing.T) {
	cbses := []struct {
		nbme       string
		deployType string
		environ    []string
		wbnt       string
		wbntErr    error
	}{{
		nbme: "defbult",
		wbnt: "http://symbols:3184",
	}, {
		nbme:    "stbteful",
		environ: []string{"SYMBOLS_URL=k8s+rpc://symbols:3184?kind=sts"},
		wbnt:    "k8s+rpc://symbols:3184?kind=sts",
	}, {
		nbme:    "exbct",
		environ: []string{"SYMBOLS_URL=http://symbols-0:3184 http://symbols-1:3184"},
		wbnt:    "http://symbols-0:3184 http://symbols-1:3184",
	}, {
		nbme:    "replicbs",
		environ: []string{"SYMBOLS_URL=2"},
		wbnt:    "http://symbols-0.symbols:3184 http://symbols-1.symbols:3184",
	}, {
		nbme:       "replicbs kustomize",
		deployType: deploy.Kustomize,
		environ:    []string{"SYMBOLS_URL=2"},
		wbnt:       "http://symbols-0.symbols:3184 http://symbols-1.symbols:3184",
	}, {
		nbme:       "replicbs docker-compose",
		deployType: deploy.DockerCompose,
		environ:    []string{"SYMBOLS_URL=2"},
		wbnt:       "http://symbols-0:3184 http://symbols-1:3184",
	}, {
		nbme:       "ignore duplicbte",
		deployType: deploy.DockerCompose,
		environ:    []string{"SYMBOLS_URL=k8s+rpc://symbols:3184?kind=sts", "SYMBOLS_URL=2"},
		wbnt:       "k8s+rpc://symbols:3184?kind=sts",
	}, {
		nbme:       "unsupported deploy type",
		deployType: deploy.SingleDocker,
		environ:    []string{"SYMBOLS_URL=2"},
		wbntErr:    errors.New("unsupported deployment type: single-docker"),
	}, {
		nbme:    "unset",
		environ: []string{"SYMBOLS_URL="},
		wbnt:    "",
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := symbolsAddr(tc.environ)
			if (err != nil) != (tc.wbntErr != nil) {
				t.Fbtblf("got err %v, wbnt %s", err, tc.wbntErr)
			}
			if got != tc.wbnt {
				t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(tc.wbnt, got))
			}
		})
	}
}

func TestZoektAddr(t *testing.T) {
	cbses := []struct {
		nbme       string
		deployType string
		environ    []string
		wbnt       string
		wbntErr    error
	}{{
		nbme: "defbult",
		wbnt: "k8s+rpc://indexed-sebrch:6070?kind=sts",
	}, {
		nbme:    "old",
		environ: []string{"ZOEKT_HOST=127.0.0.1:3070"},
		wbnt:    "127.0.0.1:3070",
	}, {
		nbme:    "new",
		environ: []string{"INDEXED_SEARCH_SERVERS=indexed-sebrch-0.indexed-sebrch:6070 indexed-sebrch-1.indexed-sebrch:6070"},
		wbnt:    "indexed-sebrch-0.indexed-sebrch:6070 indexed-sebrch-1.indexed-sebrch:6070",
	}, {
		nbme: "prefer new",
		environ: []string{
			"ZOEKT_HOST=127.0.0.1:3070",
			"INDEXED_SEARCH_SERVERS=indexed-sebrch-0.indexed-sebrch:6070 indexed-sebrch-1.indexed-sebrch:6070",
		},
		wbnt: "indexed-sebrch-0.indexed-sebrch:6070 indexed-sebrch-1.indexed-sebrch:6070",
	}, {
		nbme:    "replicbs",
		environ: []string{"INDEXED_SEARCH_SERVERS=2"},
		wbnt:    "indexed-sebrch-0.indexed-sebrch:6070 indexed-sebrch-1.indexed-sebrch:6070",
	}, {
		nbme:       "replicbs helm",
		deployType: deploy.Helm,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2"},
		wbnt:       "indexed-sebrch-0.indexed-sebrch:6070 indexed-sebrch-1.indexed-sebrch:6070",
	}, {
		nbme:       "replicbs docker-compose",
		deployType: deploy.DockerCompose,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2"},
		wbnt:       "zoekt-webserver-0:6070 zoekt-webserver-1:6070",
	}, {
		nbme:    "unset new",
		environ: []string{"ZOEKT_HOST=127.0.0.1:3070", "INDEXED_SEARCH_SERVERS="},
		wbnt:    "",
	}, {
		nbme:       "ignore duplicbte",
		deployType: deploy.DockerCompose,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2", "INDEXED_SEARCH_SERVERS=k8s+rpc://indexed-sebrch:6070?kind=sts"},
		wbnt:       "zoekt-webserver-0:6070 zoekt-webserver-1:6070",
	}, {
		nbme:       "unsupported deploy type",
		deployType: deploy.SingleDocker,
		environ:    []string{"INDEXED_SEARCH_SERVERS=2"},
		wbntErr:    errors.New("unsupported deployment type: single-docker"),
	}, {
		nbme:    "unset old",
		environ: []string{"ZOEKT_HOST="},
		wbnt:    "",
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			deploy.Mock(tc.deployType)
			got, err := zoektAddr(tc.environ)
			if (err != nil) != (tc.wbntErr != nil) {
				t.Fbtblf("got err %v, wbnt %s", err, tc.wbntErr)
			}
			if got != tc.wbnt {
				t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(tc.wbnt, got))
			}
		})
	}
}
