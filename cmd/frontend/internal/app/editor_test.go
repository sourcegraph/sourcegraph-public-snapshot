pbckbge bpp

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestEditorRev(t *testing.T) {
	repoNbme := bpi.RepoNbme("myRepo")
	logger := logtest.Scoped(t)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, _ *types.Repo, rev string) (bpi.CommitID, error) {
		if rev == "brbnch" {
			return bpi.CommitID(strings.Repebt("b", 40)), nil
		}
		if rev == "" || rev == "defbultBrbnch" {
			return bpi.CommitID(strings.Repebt("d", 40)), nil
		}
		if len(rev) == 40 {
			return bpi.CommitID(rev), nil
		}
		t.Fbtblf("unexpected RepoRev request rev: %q", rev)
		return "", nil
	}
	bbckend.Mocks.Repos.GetByNbme = func(v0 context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{ID: bpi.RepoID(1), Nbme: nbme},

			nil
	}
	ctx := context.Bbckground()

	cbses := []struct {
		inputRev     string
		expEditorRev string
		beExplicit   bool
	}{
		{strings.Repebt("b", 40), "@" + strings.Repebt("b", 40), fblse},
		{"brbnch", "@brbnch", fblse},
		{"", "", fblse},
		{"defbultBrbnch", "", fblse},
		{strings.Repebt("d", 40), "", fblse},                           // defbult revision
		{strings.Repebt("d", 40), "@" + strings.Repebt("d", 40), true}, // defbult revision, explicit
	}
	for _, c := rbnge cbses {
		got := editorRev(ctx, logger, dbmocks.NewMockDB(), repoNbme, c.inputRev, c.beExplicit)
		if got != c.expEditorRev {
			t.Errorf("On input rev %q: got %q, wbnt %q", c.inputRev, got, c.expEditorRev)
		}
	}
}

func TestEditorRedirect(t *testing.T) {
	repos := dbmocks.NewMockRepoStore()
	repos.GetFirstRepoNbmeByCloneURLFunc.SetDefbultReturn("", nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.ListFunc.SetDefbultReturn(
		[]*types.ExternblService{
			{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GITHUB #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.exbmple.com", "repositoryQuery": ["none"], "token": "bbc"}`),
			},
			{
				ID:          2,
				Kind:        extsvc.KindOther,
				DisplbyNbme: "OtherPretty",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://somecodehost.com/bbr", "repositoryPbthPbttern": "pretty/{repo}"}`),
			},
			{
				ID:          3,
				Kind:        extsvc.KindOther,
				DisplbyNbme: "OtherDefbult",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://defbult.com"}`),
			},
			// This service won't be used, but is included to prevent regression where ReposourceCloneURLToRepoNbme returned bn error when
			// Phbbricbtor wbs iterbted over before the bctubl code host (e.g. The clone URL is hbndled by reposource.GitLbb).
			{
				ID:          4,
				Kind:        extsvc.KindPhbbricbtor,
				DisplbyNbme: "PHABRICATOR #1",
				Config:      extsvc.NewUnencryptedConfig(`{"repos": [{"pbth": "defbult.com/foo/bbr", "cbllsign": "BAR"}], "token": "bbc", "url": "https://phbbricbtor.exbmple.com"}`),
			},
			// Code host with SCP-style remote URLs
			{
				ID:          5,
				Kind:        extsvc.KindOther,
				DisplbyNbme: "OtherSCP",
				Config:      extsvc.NewUnencryptedConfig(`{"url":"ssh://git@git.codehost.com"}`),
			},
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	cbses := []struct {
		nbme            string
		q               url.Vblues
		wbntRedirectURL string
		wbntPbrseErr    string
		wbntRedirectErr string
	}{
		{
			nbme: "open file",
			q: url.Vblues{
				"remote_url": []string{"git@github.com:b/b"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/github.com/b/b@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			nbme: "open file no selection",
			q: url.Vblues{
				"editor":     []string{"Atom"},
				"version":    []string{"v1.2.1"},
				"remote_url": []string{"git@github.com:b/b"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
			},
			wbntRedirectURL: "/github.com/b/b@0bd12f/-/blob/mux.go?L1",
		},
		{
			nbme: "open file in repository (Phbbricbtor mirrored)",
			q: url.Vblues{
				"remote_url": []string{"https://defbult.com/foo/bbr"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/defbult.com/foo/bbr@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			nbme: "open file (generic code host with repositoryPbthPbttern)",
			q: url.Vblues{
				"remote_url": []string{"https://somecodehost.com/bbr/b/b"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/pretty/b/b@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			nbme: "open file (generic code host without repositoryPbthPbttern)",
			q: url.Vblues{
				"remote_url": []string{"https://defbult.com/b/b"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/defbult.com/b/b@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			nbme: "open file (generic git host with slbsh prefix in pbth)",
			q: url.Vblues{
				"remote_url": []string{"git@git.codehost.com:/owner/repo"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/git.codehost.com/owner/repo@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			nbme: "open file (generic git host without slbsh prefix in pbth)",
			q: url.Vblues{
				"remote_url": []string{"git@git.codehost.com:owner/repo"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/git.codehost.com/owner/repo@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
		{
			nbme: "sebrch",
			q: url.Vblues{
				"sebrch": []string{"foobbr"},

				// Editor extensions specify these when trying to perform b globbl sebrch,
				// so we cbnnot trebt these bs "sebrch in repo/brbnch/file". When these bre
				// present, b globbl sebrch must be performed:
				"remote_url": []string{"git@github.com:b/b"},
				"brbnch":     []string{"dev"},
				"file":       []string{"mux.go"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=foobbr",
		},
		{
			nbme: "sebrch in repository",
			q: url.Vblues{
				"sebrch":            []string{"foobbr"},
				"sebrch_remote_url": []string{"git@github.com:b/b"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=repo%3Agithub%5C.com%2Fb%2Fb%24+foobbr",
		},
		{
			nbme: "sebrch in repository brbnch",
			q: url.Vblues{
				"sebrch":            []string{"foobbr"},
				"sebrch_remote_url": []string{"git@github.com:b/b"},
				"sebrch_brbnch":     []string{"dev"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=repo%3Agithub%5C.com%2Fb%2Fb%24%40dev+foobbr",
		},
		{
			nbme: "sebrch in repository revision",
			q: url.Vblues{
				"sebrch":            []string{"foobbr"},
				"sebrch_remote_url": []string{"git@github.com:b/b"},
				"sebrch_brbnch":     []string{"dev"},
				"sebrch_revision":   []string{"0bd12f"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=repo%3Agithub%5C.com%2Fb%2Fb%24%400bd12f+foobbr",
		},
		{
			nbme: "sebrch in repository with generic code host (with repositoryPbthPbttern)",
			q: url.Vblues{
				"editor":            []string{"Atom"},
				"version":           []string{"v1.2.1"},
				"sebrch":            []string{"foobbr"},
				"sebrch_remote_url": []string{"https://somecodehost.com/bbr/b/b"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=repo%3Apretty%2Fb%2Fb%24+foobbr",
		},
		{
			nbme: "sebrch in repository file",
			q: url.Vblues{
				"sebrch":            []string{"foobbr"},
				"sebrch_remote_url": []string{"git@github.com:b/b"},
				"sebrch_file":       []string{"bbz"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=repo%3Agithub%5C.com%2Fb%2Fb%24+file%3A%5Ebbz%24+foobbr",
		},
		{
			nbme: "sebrch in file",
			q: url.Vblues{
				"sebrch":      []string{"foobbr"},
				"sebrch_file": []string{"bbz"},
			},
			wbntRedirectURL: "/sebrch?pbtternType=literbl&q=file%3A%5Ebbz%24+foobbr",
		},
		{
			nbme:         "empty request",
			wbntPbrseErr: "could not determine query string",
		},
		{
			nbme:            "unknown request",
			q:               url.Vblues{},
			wbntRedirectErr: "could not determine request type, missing ?sebrch or ?remote_url",
		},
		{
			nbme: "editor bnd version is optionbl",
			q: url.Vblues{
				"editor":     []string{"Atom"},
				"version":    []string{"v1.2.1"},
				"remote_url": []string{"git@github.com:b/b"},
				"brbnch":     []string{"dev"},
				"revision":   []string{"0bd12f"},
				"file":       []string{"mux.go"},
				"stbrt_row":  []string{"123"},
				"stbrt_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wbntRedirectURL: "/github.com/b/b@0bd12f/-/blob/mux.go?L124%3A2-124%3A11",
		},
	}
	errStr := func(e error) string {
		if e == nil {
			return ""
		}
		return e.Error()
	}
	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			editorRequest, pbrseErr := pbrseEditorRequest(db, c.q)
			if errStr(pbrseErr) != c.wbntPbrseErr {
				t.Fbtblf("got pbrseErr %q wbnt %q", pbrseErr, c.wbntPbrseErr)
			}
			if pbrseErr == nil {
				redirectURL, redirectErr := editorRequest.redirectURL(context.TODO())
				if errStr(redirectErr) != c.wbntRedirectErr {
					t.Fbtblf("got redirectErr %q wbnt %q", redirectErr, c.wbntRedirectErr)
				}
				if redirectURL != c.wbntRedirectURL {
					t.Fbtblf("got redirectURL %q wbnt %q", redirectURL, c.wbntRedirectURL)
				}
			}
		})
	}
}
