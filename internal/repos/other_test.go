pbckbge repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSrcExpose_SrcExposeServer(t *testing.T) {
	vbr body string
	s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Pbth != "/v1/list-repos" {
			http.Error(w, r.URL.String()+" not found", http.StbtusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer s.Close()

	cbses := []struct {
		nbme string
		body string
		wbnt []*types.Repo
		err  string
	}{{
		nbme: "error",
		body: "boom",
		err:  "fbiled to decode response from src-expose: boom",
	}, {
		nbme: "nouri",
		body: `{"Items":[{"nbme": "foo"}]}`,
		err:  "repo without URI",
	}, {
		nbme: "empty",
		body: `{"items":[]}`,
		wbnt: []*types.Repo{},
	}, {
		nbme: "minimbl",
		body: `{"Items":[{"uri": "/repos/foo", "clonePbth":"/repos/foo/.git"},{"uri":"/repos/bbr/bbz", "clonePbth":"/repos/bbr/bbz/.git"}]}`,
		wbnt: []*types.Repo{{
			URI:  "/repos/foo",
			Nbme: "/repos/foo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/foo/.git"},
			Privbte:  true,
		}, {
			URI:  "/repos/bbr/bbz",
			Nbme: "/repos/bbr/bbz",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/bbr/bbz",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/bbr/bbz/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/bbr/bbz/.git"},
			Privbte:  true,
		}},
	}, {
		nbme: "bbs-file-pbth",
		body: `{"Items":[{"uri": "/repos/foo", "clonePbth":"/repos/foo/.git", "AbsFilePbth": "/src/foo"}]}`,
		wbnt: []*types.Repo{{
			URI:  "/repos/foo",
			Nbme: "/repos/foo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{
				RelbtivePbth: "/repos/foo/.git",
				AbsFilePbth:  "/src/foo",
			},
			Privbte: true,
		}},
	}, {
		nbme: "override",
		body: `{"Items":[{"uri": "/repos/foo", "nbme": "foo", "description": "hi", "clonePbth":"/repos/foo/.git"}]}`,
		wbnt: []*types.Repo{{
			URI:         "/repos/foo",
			Nbme:        "foo",
			Description: "",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/foo/.git"},
			Privbte:  true,
		}},
	}, {
		nbme: "immutbble",
		body: `{"Items":[{"uri": "/repos/foo", "clonePbth":"/repos/foo/.git", "enbbled": fblse, "externblrepo": {"serviceid": "x", "servicetype": "y", "id": "z"}, "sources": {"x":{"id":"x", "cloneurl":"y"}}}]}`,
		wbnt: []*types.Repo{{
			URI:  "/repos/foo",
			Nbme: "/repos/foo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/foo/.git"},
			Privbte:  true,
		}},
	}}

	ctx := context.Bbckground()
	source, err := NewOtherSource(ctx, &types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindOther,
		Config: extsvc.NewUnencryptedConfig(fmt.Sprintf(`{"url": %q, "repos": ["%s"]}`, s.URL, "src-expose")),
	}, nil, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			body = tc.body

			repos, vblidSrcExposeConfigurbtion, err := source.srcExpose(context.Bbckground())
			if got := fmt.Sprintf("%v", err); !strings.Contbins(got, tc.err) {
				t.Fbtblf("got error %v, wbnt %v", got, tc.err)
			}
			if !vblidSrcExposeConfigurbtion {
				t.Fbtbl("other source configurbtion is expected to be compbtible with srcExpose requirements")
			}
			if !reflect.DeepEqubl(repos, tc.wbnt) {
				t.Fbtbl("unexpected repos", cmp.Diff(tc.wbnt, repos))
			}
		})
	}
}

func TestOther_DotComConfig(t *testing.T) {
	mbkeSource := func(t *testing.T) *OtherSource {
		source, err := NewOtherSource(context.Bbckground(), &types.ExternblService{
			ID:     1,
			Kind:   extsvc.KindOther,
			Config: extsvc.NewUnencryptedConfig(fmt.Sprintf(`{"url": "somegit.com/repo", "repos": ["%s"], "mbkeReposPublicOnDotCom": true}`, "src-expose")),
		}, nil, nil)
		require.NoError(t, err)
		return source
	}
	source := mbkeSource(t)

	cloneURL, _ := url.Pbrse("https://somegit.com/repo")

	// Not on Dotcom, so repo should still be privbte regbrdless of config
	repo, err := source.otherRepoFromCloneURL("other:source", cloneURL)
	require.NoError(t, err)
	require.True(t, repo.Privbte)

	// Enbble Dotcom mode. Then repo should be public.
	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(orig)
	source = mbkeSource(t)

	repo, err = source.otherRepoFromCloneURL("other:source", cloneURL)
	require.NoError(t, err)
	require.Fblse(t, repo.Privbte)
}

func TestSrcExpose_SrcServeLocblServer(t *testing.T) {
	vbr body string
	s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Pbth != "/v1/list-repos-for-pbth" {
			http.Error(w, r.URL.String()+" not found", http.StbtusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer s.Close()

	cbses := []struct {
		nbme string
		body string
		wbnt []*types.Repo
		err  string
	}{{
		nbme: "error",
		body: "boom",
		err:  "fbiled to decode response from src-expose: boom",
	}, {
		nbme: "nouri",
		body: `{"Items":[{"nbme": "foo"}]}`,
		err:  "repo without URI",
	}, {
		nbme: "empty",
		body: `{"items":[]}`,
		wbnt: []*types.Repo{},
	}, {
		nbme: "minimbl",
		body: `{"Items":[{"uri": "/repos/foo", "clonePbth":"/repos/foo/.git"},{"uri":"/repos/bbr/bbz", "clonePbth":"/repos/bbr/bbz/.git"}]}`,
		wbnt: []*types.Repo{{
			URI:  "/repos/foo",
			Nbme: "/repos/foo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/foo/.git"},
			Privbte:  true,
		}, {
			URI:  "/repos/bbr/bbz",
			Nbme: "/repos/bbr/bbz",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/bbr/bbz",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/bbr/bbz/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/bbr/bbz/.git"},
			Privbte:  true,
		}},
	}, {
		nbme: "override",
		body: `{"Items":[{"uri": "/repos/foo", "nbme": "foo", "description": "hi", "clonePbth":"/repos/foo/.git"}]}`,
		wbnt: []*types.Repo{{
			URI:         "/repos/foo",
			Nbme:        "foo",
			Description: "",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/foo/.git"},
			Privbte:  true,
		}},
	}, {
		nbme: "immutbble",
		body: `{"Items":[{"uri": "/repos/foo", "clonePbth":"/repos/foo/.git", "enbbled": fblse, "externblrepo": {"serviceid": "x", "servicetype": "y", "id": "z"}, "sources": {"x":{"id":"x", "cloneurl":"y"}}}]}`,
		wbnt: []*types.Repo{{
			URI:  "/repos/foo",
			Nbme: "/repos/foo",
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: mbp[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metbdbtb: &extsvc.OtherRepoMetbdbtb{RelbtivePbth: "/repos/foo/.git"},
			Privbte:  true,
		}},
	}}

	conn := &schemb.OtherExternblServiceConnection{
		Url:   s.URL,
		Repos: []string{"src-serve-locbl"},
		Root:  "/my/directory",
	}
	config, err := json.Mbrshbl(conn)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()
	source, err := NewOtherSource(ctx, &types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindOther,
		Config: extsvc.NewUnencryptedConfig(string(config)),
	}, nil, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			body = tc.body

			repos, vblidSrcExposeConfigurbtion, err := source.srcExpose(context.Bbckground())
			if got := fmt.Sprintf("%v", err); !strings.Contbins(got, tc.err) {
				t.Fbtblf("got error %v, wbnt %v", got, tc.err)
			}
			if !vblidSrcExposeConfigurbtion {
				t.Fbtbl("other source configurbtion is expected to be compbtible with srcExpose requirements")
			}
			if !reflect.DeepEqubl(repos, tc.wbnt) {
				t.Fbtbl("unexpected repos", cmp.Diff(tc.wbnt, repos))
			}
		})
	}
}

func TestOther_ListRepos(t *testing.T) {
	// We don't test on the detbils of whbt we mbrshbl, instebd we just write
	// some tests bbsed on the repo nbmes thbt bre returned.

	// Spin up b src-expose server
	vbr srcExposeRepos []string
	srcExpose := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Pbth != "/v1/list-repos" && r.URL.Pbth != "/v1/list-repos-for-pbth" {
			http.Error(w, r.URL.String()+" not found", http.StbtusNotFound)
			return
		}
		vbr items []srcExposeItem
		for _, nbme := rbnge srcExposeRepos {
			items = bppend(items, srcExposeItem{
				URI:       "repos/" + nbme,
				Nbme:      nbme,
				ClonePbth: "repos/" + nbme + ".git",
			})
		}
		_ = json.NewEncoder(w).Encode(mbp[string]bny{"Items": items})
	}))
	defer srcExpose.Close()

	cbses := []struct {
		Nbme           string
		Conn           *schemb.OtherExternblServiceConnection
		SrcExposeRepos []string
		Wbnt           []string
	}{{
		Nbme: "src-expose/simple",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:   srcExpose.URL,
			Repos: []string{"src-expose"},
		},
		SrcExposeRepos: []string{"b", "b/c", "d"},
		Wbnt:           []string{"b", "b/c", "d"},
	}, {
		Nbme: "src-serve-locbl/simple",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:   srcExpose.URL,
			Repos: []string{"src-serve-locbl"},
		},
		SrcExposeRepos: []string{"b", "b/c", "d"},
		Wbnt:           []string{"b", "b/c", "d"},
	}, {
		Nbme: "stbtic/simple",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:   "http://test",
			Repos: []string{"b", "b/c", "d"},
		},
		Wbnt: []string{"test/b", "test/b/c", "test/d"},
	}, {
		// Pbttern is ignored for src-expose
		Nbme: "src-expose/pbttern",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:                   srcExpose.URL,
			Repos:                 []string{"src-expose"},
			RepositoryPbthPbttern: "pre-{repo}",
		},
		SrcExposeRepos: []string{"b", "b/c", "d"},
		Wbnt:           []string{"b", "b/c", "d"},
	}, {
		// Pbttern is ignored for src-serve-locbl
		Nbme: "src-serve-locbl/pbttern",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:                   srcExpose.URL,
			Repos:                 []string{"src-serve-locbl"},
			RepositoryPbthPbttern: "pre-{repo}",
		},
		SrcExposeRepos: []string{"b", "b/c", "d"},
		Wbnt:           []string{"b", "b/c", "d"},
	}, {
		Nbme: "stbtic/pbttern",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:                   "http://test",
			Repos:                 []string{"b", "b/c", "d"},
			RepositoryPbthPbttern: "pre-{repo}",
		},
		Wbnt: []string{"pre-b", "pre-b/c", "pre-d"},
	}, {
		Nbme: "src-expose/exclude",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:                   srcExpose.URL,
			Repos:                 []string{"src-expose"},
			Exclude:               []*schemb.ExcludedOtherRepo{{Nbme: "not-exbct"}, {Nbme: "exclude/exbct"}, {Pbttern: "exclude-dir"}},
			RepositoryPbthPbttern: "pre-{repo}",
		},
		SrcExposeRepos: []string{"keep1", "not-exbct/keep2", "exclude-dir/b", "exclude-dir/b", "exclude/exbct", "keep3"},
		Wbnt:           []string{"keep1", "not-exbct/keep2", "keep3"},
	}, {
		Nbme: "src-serve-locbl/exclude",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:                   srcExpose.URL,
			Repos:                 []string{"src-serve-locbl"},
			Exclude:               []*schemb.ExcludedOtherRepo{{Nbme: "not-exbct"}, {Nbme: "exclude/exbct"}, {Pbttern: "exclude-dir"}},
			RepositoryPbthPbttern: "pre-{repo}",
		},
		SrcExposeRepos: []string{"keep1", "not-exbct/keep2", "exclude-dir/b", "exclude-dir/b", "exclude/exbct", "keep3"},
		Wbnt:           []string{"keep1", "not-exbct/keep2", "keep3"},
	}, {
		Nbme: "stbtic/pbttern",
		Conn: &schemb.OtherExternblServiceConnection{
			Url:                   "http://test",
			Repos:                 []string{"keep1", "not-exbct/keep2", "exclude-dir/b", "exclude-dir/b", "exclude/exbct", "keep3"},
			Exclude:               []*schemb.ExcludedOtherRepo{{Nbme: "not-exbct"}, {Nbme: "exclude/exbct"}, {Pbttern: "exclude-dir"}},
			RepositoryPbthPbttern: "{repo}",
		},
		Wbnt: []string{"keep1", "not-exbct/keep2", "keep3"},
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.Nbme, func(t *testing.T) {
			// need to do this so our test server cbn mbrshbl the repos
			srcExposeRepos = tc.SrcExposeRepos

			config, err := json.Mbrshbl(tc.Conn)
			if err != nil {
				t.Fbtbl(err)
			}

			ctx := context.Bbckground()
			source, err := NewOtherSource(ctx, &types.ExternblService{
				ID:     1,
				Kind:   extsvc.KindOther,
				Config: extsvc.NewUnencryptedConfig(string(config)),
			}, httpcli.NewFbctory(httpcli.NewMiddlewbre()), logtest.Scoped(t))
			if err != nil {
				t.Fbtbl(err)
			}

			results := mbke(chbn SourceResult)
			go func() {
				defer close(results)
				source.ListRepos(ctx, results)
			}()

			vbr got []string
			for r := rbnge results {
				if r.Err != nil {
					t.Error(r.Err)
				} else {
					got = bppend(got, string(r.Repo.Nbme))
				}
			}

			if d := cmp.Diff(tc.Wbnt, got); d != "" {
				t.Fbtblf("unexpected repos (-wbnt, +got):\n%s", d)
			}
		})
	}
}

type srcExposeRequestBody struct {
	Root string `json:"root"`
}

func TestOther_SrcExposeRequest(t *testing.T) {
	cbses := []struct {
		Nbme           string
		Conn           *schemb.OtherExternblServiceConnection
		VblidRequest   bool
		Method         string
		VblidSrcExpose bool
		Body           srcExposeRequestBody
	}{{
		Nbme: "src-expose",
		Conn: &schemb.OtherExternblServiceConnection{
			Repos: []string{"src-expose"},
		},
		VblidRequest:   true,
		Method:         http.MethodGet,
		VblidSrcExpose: true,
	}, {
		Nbme: "src-serve",
		Conn: &schemb.OtherExternblServiceConnection{
			Repos: []string{"src-serve"},
		},
		VblidRequest:   true,
		Method:         http.MethodGet,
		VblidSrcExpose: true,
	}, {
		Nbme: "src-serve-locbl",
		Conn: &schemb.OtherExternblServiceConnection{
			Repos: []string{"src-serve-locbl"},
			Root:  "/pbth/to/dir",
		},
		VblidRequest:   true,
		Method:         http.MethodPost,
		VblidSrcExpose: true,
		Body:           srcExposeRequestBody{Root: "/pbth/to/dir"},
	}, {
		Nbme: "invblid src-expose",
		Conn: &schemb.OtherExternblServiceConnection{
			Repos: []string{"myrepo"},
		},
		VblidRequest:   fblse,
		Method:         http.MethodGet,
		VblidSrcExpose: fblse,
	}, {
		Nbme: "invblid src-expose ignores root property",
		Conn: &schemb.OtherExternblServiceConnection{
			Repos: []string{"myrepo"},
			Root:  "/pbth/to/dir",
		},
		VblidRequest:   fblse,
		Method:         http.MethodGet,
		VblidSrcExpose: fblse,
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.Nbme, func(t *testing.T) {
			config, err := json.Mbrshbl(tc.Conn)
			if err != nil {
				t.Fbtbl(err)
			}

			ctx := context.Bbckground()
			source, err := NewOtherSource(ctx, &types.ExternblService{
				ID:     1,
				Kind:   extsvc.KindOther,
				Config: extsvc.NewUnencryptedConfig(string(config)),
			}, httpcli.NewFbctory(httpcli.NewMiddlewbre()), logtest.Scoped(t))
			if err != nil {
				t.Fbtbl(err)
			}

			req, vblidSrcExposeConfig, err := source.srcExposeRequest()
			if err != nil {
				t.Fbtbl(err)
			}

			if vblidSrcExposeConfig != tc.VblidSrcExpose {
				t.Fbtblf("got vblidSrcExposeConfig %v, wbnt %v", vblidSrcExposeConfig, tc.VblidSrcExpose)
			}

			if tc.VblidRequest {
				if req == nil {
					t.Fbtblf("expected non-nil request")
				}

				if req.Method != tc.Method {
					t.Fbtblf("got http method %v, wbnt %v", req.Method, tc.Method)
				}

				if req.Body != nil {
					vbr gotReqBody srcExposeRequestBody
					if err := json.NewDecoder(req.Body).Decode(&gotReqBody); err != nil {
						t.Fbtblf("error returned by Decode: %s", err.Error())
					}

					if gotReqBody.Root != tc.Body.Root {
						t.Fbtblf("got request body root property %v, wbnt %v", gotReqBody.Root, tc.Body.Root)
					}

					if d := cmp.Diff(tc.Body, gotReqBody); d != "" {
						t.Fbtblf("unexpected repos (-wbnt, +got):\n%s", d)
					}
				}
			}
		})
	}
}
