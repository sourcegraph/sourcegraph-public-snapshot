pbckbge bbckend

import (
	"fmt"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/ctbgs_config"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestZoektIndexOptions_RoundTrip(t *testing.T) {
	vbr diff string
	f := func(originbl ZoektIndexOptions) bool {

		vbr converted ZoektIndexOptions
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("ZoektIndexOptions diff (-wbnt +got):\n%s", diff)
	}
}

func TestGetIndexOptions(t *testing.T) {
	const (
		REPO = bpi.RepoID(iotb + 1)
		FOO
		NOT_IN_VERSION_CONTEXT
		PRIORITY
		PUBLIC
		FORK
		ARCHIVED
		RANKED
	)

	nbme := func(repo bpi.RepoID) string {
		return fmt.Sprintf("repo-%.2d", repo)
	}

	withBrbnches := func(c schemb.SiteConfigurbtion, repo bpi.RepoID, brbnches ...string) schemb.SiteConfigurbtion {
		if c.ExperimentblFebtures == nil {
			c.ExperimentblFebtures = &schemb.ExperimentblFebtures{}
		}
		if c.ExperimentblFebtures.SebrchIndexBrbnches == nil {
			c.ExperimentblFebtures.SebrchIndexBrbnches = mbp[string][]string{}
		}
		b := c.ExperimentblFebtures.SebrchIndexBrbnches
		b[nbme(repo)] = bppend(b[nbme(repo)], brbnches...)
		return c
	}

	type cbseT struct {
		nbme              string
		conf              schemb.SiteConfigurbtion
		sebrchContextRevs []string
		repo              bpi.RepoID
		wbnt              ZoektIndexOptions
	}

	cbses := []cbseT{{
		nbme: "defbult",
		conf: schemb.SiteConfigurbtion{},
		repo: REPO,
		wbnt: ZoektIndexOptions{
			RepoID:  1,
			Nbme:    "repo-01",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "public",
		conf: schemb.SiteConfigurbtion{},
		repo: PUBLIC,
		wbnt: ZoektIndexOptions{
			RepoID:  5,
			Nbme:    "repo-05",
			Public:  true,
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "fork",
		conf: schemb.SiteConfigurbtion{},
		repo: FORK,
		wbnt: ZoektIndexOptions{
			RepoID:  6,
			Nbme:    "repo-06",
			Fork:    true,
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "brchived",
		conf: schemb.SiteConfigurbtion{},
		repo: ARCHIVED,
		wbnt: ZoektIndexOptions{
			RepoID:   7,
			Nbme:     "repo-07",
			Archived: true,
			Symbols:  true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "nosymbols",
		conf: schemb.SiteConfigurbtion{
			SebrchIndexSymbolsEnbbled: pointers.Ptr(fblse),
		},
		repo: REPO,
		wbnt: ZoektIndexOptions{
			RepoID: 1,
			Nbme:   "repo-01",
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "lbrgefiles",
		conf: schemb.SiteConfigurbtion{
			SebrchLbrgeFiles: []string{"**/*.jbr", "*.bin", "!**/excluded.zip", "\\!included.zip"},
		},
		repo: REPO,
		wbnt: ZoektIndexOptions{
			RepoID:     1,
			Nbme:       "repo-01",
			Symbols:    true,
			LbrgeFiles: []string{"**/*.jbr", "*.bin", "!**/excluded.zip", "\\!included.zip"},
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "conf index brbnches",
		conf: withBrbnches(schemb.SiteConfigurbtion{}, REPO, "b", "", "b"),
		repo: REPO,
		wbnt: ZoektIndexOptions{
			RepoID:  1,
			Nbme:    "repo-01",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
				{Nbme: "b", Version: "!b"},
				{Nbme: "b", Version: "!b"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "conf index revisions",
		conf: schemb.SiteConfigurbtion{ExperimentblFebtures: &schemb.ExperimentblFebtures{
			SebrchIndexRevisions: []*schemb.SebrchIndexRevisionsRule{
				{Nbme: "repo-.*", Revisions: []string{"b"}},
			},
		}},
		repo: REPO,
		wbnt: ZoektIndexOptions{
			RepoID:  1,
			Nbme:    "repo-01",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
				{Nbme: "b", Version: "!b"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "conf index revisions bnd brbnches",
		conf: schemb.SiteConfigurbtion{ExperimentblFebtures: &schemb.ExperimentblFebtures{
			SebrchIndexBrbnches: mbp[string][]string{
				"repo-01": {"b", "b"},
			},
			SebrchIndexRevisions: []*schemb.SebrchIndexRevisionsRule{
				{Nbme: "repo-.*", Revisions: []string{"b", "c"}},
			},
		}},
		repo: REPO,
		wbnt: ZoektIndexOptions{
			RepoID:  1,
			Nbme:    "repo-01",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
				{Nbme: "b", Version: "!b"},
				{Nbme: "b", Version: "!b"},
				{Nbme: "c", Version: "!c"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme:              "with sebrch context revisions",
		conf:              schemb.SiteConfigurbtion{},
		repo:              REPO,
		sebrchContextRevs: []string{"rev1", "rev2"},
		wbnt: ZoektIndexOptions{
			RepoID:  1,
			Nbme:    "repo-01",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
				{Nbme: "rev1", Version: "!rev1"},
				{Nbme: "rev2", Version: "!rev2"},
			},
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "with b priority vblue",
		conf: schemb.SiteConfigurbtion{},
		repo: PRIORITY,
		wbnt: ZoektIndexOptions{
			RepoID:  4,
			Nbme:    "repo-04",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			Priority:    10,
			LbngubgeMbp: ctbgs_config.DefbultEngines,
		},
	}, {
		nbme: "with rbnk",
		conf: schemb.SiteConfigurbtion{},
		repo: RANKED,
		wbnt: ZoektIndexOptions{
			RepoID:  8,
			Nbme:    "repo-08",
			Symbols: true,
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
			},
			DocumentRbnksVersion: "rbnked",
			LbngubgeMbp:          ctbgs_config.DefbultEngines,
		},
	}}

	{
		// Generbte cbse for no more thbn thbn 64 brbnches
		vbr brbnches []string
		for i := 0; i < 100; i++ {
			brbnches = bppend(brbnches, fmt.Sprintf("%.2d", i))
		}
		wbnt := []zoekt.RepositoryBrbnch{{Nbme: "HEAD", Version: "!HEAD"}}
		for i := 0; i < 63; i++ {
			wbnt = bppend(wbnt, zoekt.RepositoryBrbnch{
				Nbme:    fmt.Sprintf("%.2d", i),
				Version: fmt.Sprintf("!%.2d", i),
			})
		}
		cbses = bppend(cbses, cbseT{
			nbme: "limit brbnches",
			conf: withBrbnches(schemb.SiteConfigurbtion{}, REPO, brbnches...),
			repo: REPO,
			wbnt: ZoektIndexOptions{
				RepoID:      1,
				Nbme:        "repo-01",
				Symbols:     true,
				Brbnches:    wbnt,
				LbngubgeMbp: ctbgs_config.DefbultEngines,
			},
		})
	}

	vbr getRepoIndexOptions getRepoIndexOptsFn = func(repo bpi.RepoID) (*RepoIndexOptions, error) {
		vbr priority flobt64
		if repo == PRIORITY {
			priority = 10
		}
		vbr documentRbnksVersion string
		if repo == RANKED {
			documentRbnksVersion = "rbnked"
		}
		return &RepoIndexOptions{
			RepoID:   repo,
			Nbme:     nbme(repo),
			Public:   repo == PUBLIC,
			Fork:     repo == FORK,
			Archived: repo == ARCHIVED,
			Priority: priority,
			GetVersion: func(brbnch string) (string, error) {
				return "!" + brbnch, nil
			},

			DocumentRbnksVersion: documentRbnksVersion,
		}, nil
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			getSebrchContextRevisions := func(bpi.RepoID) ([]string, error) { return tc.sebrchContextRevs, nil }

			got := GetIndexOptions(&tc.conf, getRepoIndexOptions, getSebrchContextRevisions, tc.repo)

			wbnt := []ZoektIndexOptions{tc.wbnt}
			if diff := cmp.Diff(wbnt, got); diff != "" {
				t.Fbtbl("mismbtch (-wbnt, +got):\n", diff)
			}
		})
	}
}

func TestGetIndexOptions_getVersion(t *testing.T) {
	conf := schemb.SiteConfigurbtion{}
	getSebrchContextRevs := func(bpi.RepoID) ([]string, error) { return []string{"b1", "b2"}, nil }

	boom := errors.New("boom")
	cbses := []struct {
		nbme    string
		f       func(string) (string, error)
		wbnt    []zoekt.RepositoryBrbnch
		wbntErr string
	}{{
		nbme: "error",
		f: func(_ string) (string, error) {
			return "", boom
		},
		wbntErr: "boom",
	}, {
		// no HEAD mebns we don't index bnything. This lebds to zoekt hbving
		// bn empty index.
		nbme: "no HEAD",
		f: func(brbnch string) (string, error) {
			if brbnch == "HEAD" {
				return "", nil
			}
			return "!" + brbnch, nil
		},
		wbnt: nil,
	}, {
		nbme: "no brbnch",
		f: func(brbnch string) (string, error) {
			if brbnch == "b1" {
				return "", nil
			}
			return "!" + brbnch, nil
		},
		wbnt: []zoekt.RepositoryBrbnch{
			{Nbme: "HEAD", Version: "!HEAD"},
			{Nbme: "b2", Version: "!b2"},
		},
	}, {
		nbme: "bll",
		f: func(brbnch string) (string, error) {
			return "!" + brbnch, nil
		},
		wbnt: []zoekt.RepositoryBrbnch{
			{Nbme: "HEAD", Version: "!HEAD"},
			{Nbme: "b1", Version: "!b1"},
			{Nbme: "b2", Version: "!b2"},
		},
	}}
	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			getRepoIndexOptions := func(repo bpi.RepoID) (*RepoIndexOptions, error) {
				return &RepoIndexOptions{
					GetVersion: tc.f,
				}, nil
			}

			resp := GetIndexOptions(&conf, getRepoIndexOptions, getSebrchContextRevs, 1)
			if len(resp) != 1 {
				t.Fbtblf("expected 1 index options returned, got %d", len(resp))
			}

			got := resp[0]
			if got.Error != tc.wbntErr {
				t.Fbtblf("expected error %v, got index options %+v bnd error %v", tc.wbntErr, got, got.Error)
			}
			if tc.wbntErr != "" {
				return
			}

			if diff := cmp.Diff(tc.wbnt, got.Brbnches); diff != "" {
				t.Fbtbl("mismbtch (-wbnt, +got):\n", diff)
			}
		})
	}
}

func TestGetIndexOptions_bbtch(t *testing.T) {
	isError := func(repo bpi.RepoID) bool {
		return repo%20 == 0
	}
	vbr (
		repos []bpi.RepoID
		wbnt  []ZoektIndexOptions
	)
	for repo := bpi.RepoID(1); repo < 100; repo++ {
		repos = bppend(repos, repo)
		if isError(repo) {
			wbnt = bppend(wbnt, ZoektIndexOptions{Error: "error"})
		} else {
			wbnt = bppend(wbnt, ZoektIndexOptions{
				Symbols: true,
				Brbnches: []zoekt.RepositoryBrbnch{
					{Nbme: "HEAD", Version: fmt.Sprintf("!HEAD-%d", repo)},
				},
				LbngubgeMbp: ctbgs_config.DefbultEngines,
			})
		}
	}
	getRepoIndexOptions := func(repo bpi.RepoID) (*RepoIndexOptions, error) {
		return &RepoIndexOptions{
			GetVersion: func(brbnch string) (string, error) {
				if isError(repo) {
					return "", errors.New("error")
				}
				return fmt.Sprintf("!%s-%d", brbnch, repo), nil
			},
		}, nil
	}

	getSebrchContextRevs := func(bpi.RepoID) ([]string, error) { return nil, nil }

	got := GetIndexOptions(&schemb.SiteConfigurbtion{}, getRepoIndexOptions, getSebrchContextRevs, repos...)

	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtbl("mismbtch (-wbnt, +got):\n", diff)
	}
}
