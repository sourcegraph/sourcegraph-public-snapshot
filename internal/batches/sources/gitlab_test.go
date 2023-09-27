pbckbge sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Mbsterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/inconshrevebble/log15"
	"github.com/stretchr/testify/bssert"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/versions"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr mockVersion = semver.MustPbrse("12.0.0-pre")
vbr mockVersion2 = semver.MustPbrse("14.10.0-pre")

func TestGitLbbSource(t *testing.T) {
	t.Run("determineVersion", func(t *testing.T) {
		t.Run("externbl service is cbched in redis", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)

			p.mockGetVersions(mockVersion.String(), p.source.client.Urn())
			v, err := p.source.determineVersion(p.ctx)

			bssert.NoError(t, err)
			bssert.True(t, mockVersion.Equbl(v), fmt.Sprintf("expected: %s, got: %s", v.String(), mockVersion.String()))
			bssert.Fblse(t, p.isGetVersionCblled, "Client.GetVersion should not be cblled")
		})

		t.Run("externbl service (mbtching the key) does not exist in redis", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)

			p.mockGetVersions("", "rbndom-urn-key-thbt-doesnt exist")
			p.mockGetVersion(mockVersion.String())
			v, err := p.source.determineVersion(p.ctx)

			bssert.NoError(t, err)
			bssert.True(t, mockVersion.Equbl(v), fmt.Sprintf("expected: %s, got: %s", v.String(), mockVersion.String()))
			bssert.True(t, p.isGetVersionCblled, "Client.GetVersion should be cblled")
		})
	})

	t.Run("CrebteDrbftChbngeset", func(t *testing.T) {
		t.Run("GitLbb version is grebter thbn 14.0.0", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)

			p.mockGetVersions(mockVersion2.String(), p.source.client.Urn())
			p.mockCrebteMergeRequest(gitlbb.CrebteMergeRequestOpts{
				SourceBrbnch: p.mr.SourceBrbnch,
				TbrgetBrbnch: p.mr.TbrgetBrbnch,
			}, p.mr, nil)
			p.mockGetMergeRequestNotes(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(p.mr.IID, nil, 20, nil)

			exists, err := p.source.CrebteDrbftChbngeset(p.ctx, p.chbngeset)
			bssert.NoError(t, err)
			bssert.True(t, strings.HbsPrefix(p.chbngeset.Title, "Drbft:"))
			bssert.Fblse(t, exists)
		})

		t.Run("GitLbb Version is less thbn 14.0.0", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)

			p.mockGetVersions(mockVersion.String(), p.source.client.Urn())
			p.mockCrebteMergeRequest(gitlbb.CrebteMergeRequestOpts{
				SourceBrbnch: p.mr.SourceBrbnch,
				TbrgetBrbnch: p.mr.TbrgetBrbnch,
			}, p.mr, nil)
			p.mockGetMergeRequestNotes(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(p.mr.IID, nil, 20, nil)

			exists, err := p.source.CrebteDrbftChbngeset(p.ctx, p.chbngeset)
			bssert.NoError(t, err)
			bssert.True(t, strings.HbsPrefix(p.chbngeset.Title, "WIP:"))
			bssert.Fblse(t, exists)
		})
	})
}

// TestGitLbbSource_ChbngesetSource tests the vbrious Chbngeset functions thbt
// implement the ChbngesetSource interfbce.
func TestGitLbbSource_ChbngesetSource(t *testing.T) {
	t.Run("CrebteChbngeset", func(t *testing.T) {
		t.Run("invblid metbdbtb", func(t *testing.T) {
			defer func() { _ = recover() }()

			p := newGitLbbChbngesetSourceTestProvider(t)
			repo := &types.Repo{Metbdbtb: struct{}{}}
			_, _ = p.source.CrebteChbngeset(p.ctx, &Chbngeset{
				RemoteRepo: repo,
				TbrgetRepo: repo,
			})
			t.Error("invblid metbdbtb did not pbnic")
		})

		t.Run("error from CrebteMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.mockCrebteMergeRequest(gitlbb.CrebteMergeRequestOpts{
				SourceBrbnch: p.mr.SourceBrbnch,
				TbrgetBrbnch: p.mr.TbrgetBrbnch,
			}, nil, inner)

			exists, hbve := p.source.CrebteChbngeset(p.ctx, p.chbngeset)
			if exists {
				t.Errorf("unexpected exists vblue: %v", exists)
			}
			if !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
		})

		t.Run("error from GetOpenMergeRequestByRefs", func(t *testing.T) {
			inner := errors.New("foo")

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.mockCrebteMergeRequest(gitlbb.CrebteMergeRequestOpts{
				SourceBrbnch: p.mr.SourceBrbnch,
				TbrgetBrbnch: p.mr.TbrgetBrbnch,
			}, nil, gitlbb.ErrMergeRequestAlrebdyExists)
			p.mockGetOpenMergeRequestByRefs(nil, inner)

			exists, hbve := p.source.CrebteChbngeset(p.ctx, p.chbngeset)
			if !exists {
				t.Errorf("unexpected exists vblue: %v", exists)
			}
			if !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
		})

		t.Run("merge request blrebdy exists", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)
			p.mockCrebteMergeRequest(gitlbb.CrebteMergeRequestOpts{
				SourceBrbnch: p.mr.SourceBrbnch,
				TbrgetBrbnch: p.mr.TbrgetBrbnch,
			}, nil, gitlbb.ErrMergeRequestAlrebdyExists)
			p.mockGetMergeRequestNotes(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(p.mr.IID, nil, 20, nil)
			p.mockGetOpenMergeRequestByRefs(p.mr, nil)

			exists, err := p.source.CrebteChbngeset(p.ctx, p.chbngeset)
			if !exists {
				t.Errorf("unexpected exists vblue: %v", exists)
			}
			if err != nil {
				t.Errorf("unexpected non-nil err: %+v", err)
			}

			if p.chbngeset.Chbngeset.Metbdbtb != p.mr {
				t.Errorf("unexpected metbdbtb: hbve %+v; wbnt %+v", p.chbngeset.Chbngeset.Metbdbtb, p.mr)
			}
		})

		t.Run("merge request is new", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)
			p.mockCrebteMergeRequest(gitlbb.CrebteMergeRequestOpts{
				SourceBrbnch: p.mr.SourceBrbnch,
				TbrgetBrbnch: p.mr.TbrgetBrbnch,
			}, p.mr, nil)
			p.mockGetMergeRequestNotes(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(p.mr.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(p.mr.IID, nil, 20, nil)

			exists, err := p.source.CrebteChbngeset(p.ctx, p.chbngeset)
			if exists {
				t.Errorf("unexpected exists vblue: %v", exists)
			}
			if err != nil {
				t.Errorf("unexpected non-nil err: %+v", err)
			}

			if p.chbngeset.Chbngeset.Metbdbtb != p.mr {
				t.Errorf("unexpected metbdbtb: hbve %+v; wbnt %+v", p.chbngeset.Chbngeset.Metbdbtb, p.mr)
			}
		})

		t.Run("integrbtion", func(t *testing.T) {
			// Repository used: https://gitlbb.com/bbtch-chbnges-testing/bbtch-chbnges-test-repo
			// This repository does not hbve bny project setting to delete source brbnches
			// butombticblly on PR merge.
			//
			// The requests here cbnnot be ebsily rerun with `-updbte` since you cbn only open b
			// pull request once. To updbte, push b new brbnch with bt lebst one commit to
			// bbtch-chbnges-test-repo, then updbte the brbnch nbmes in the test cbses below.
			//
			// You cbn updbte just this test with `-updbte GitLbbSource_CrebteChbngeset`.
			repo := &types.Repo{
				Metbdbtb: &gitlbb.Project{
					// https://gitlbb.com/bbtch-chbnges-testing/bbtch-chbnges-test-repo
					ProjectCommon: gitlbb.ProjectCommon{ID: 40370047},
				},
			}

			testCbses := []struct {
				nbme               string
				cs                 *Chbngeset
				removeSourceBrbnch bool
			}{
				{
					nbme: "no-remove-source-brbnch",
					cs: &Chbngeset{
						Title:      "This is b test PR",
						Body:       "This is the description of the test PR",
						HebdRef:    "refs/hebds/test-pr-3",
						BbseRef:    "refs/hebds/mbin",
						RemoteRepo: repo,
						TbrgetRepo: repo,
						Chbngeset:  &btypes.Chbngeset{},
					},
					removeSourceBrbnch: fblse,
				},
				{
					nbme: "yes-remove-source-brbnch",
					cs: &Chbngeset{
						Title:      "This is b test PR",
						Body:       "This is the description of the test PR",
						HebdRef:    "refs/hebds/test-pr-4",
						BbseRef:    "refs/hebds/mbin",
						RemoteRepo: repo,
						TbrgetRepo: repo,
						Chbngeset:  &btypes.Chbngeset{},
					},
					removeSourceBrbnch: true,
				},
			}

			for _, tc := rbnge testCbses {
				tc := tc
				tc.nbme = "GitLbbSource_CrebteChbngeset_" + tc.nbme

				t.Run(tc.nbme, func(t *testing.T) {
					cf, sbve := newClientFbctory(t, tc.nbme)
					defer sbve(t)

					if tc.removeSourceBrbnch {
						conf.Mock(&conf.Unified{
							SiteConfigurbtion: schemb.SiteConfigurbtion{
								BbtchChbngesAutoDeleteBrbnch: true,
							},
						})
						defer conf.Mock(nil)
					}

					lg := log15.New()
					lg.SetHbndler(log15.DiscbrdHbndler())

					svc := &types.ExternblService{
						Kind: extsvc.KindGitLbb,
						Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitLbbConnection{
							Url:   "https://gitlbb.com",
							Token: os.Getenv("GITLAB_TOKEN"),
						})),
					}

					ctx := context.Bbckground()
					gitlbbSource, err := NewGitLbbSource(ctx, svc, cf)
					if err != nil {
						t.Fbtbl(err)
					}

					_, err = gitlbbSource.CrebteChbngeset(ctx, tc.cs)
					if err != nil {
						t.Fbtbl(err)
					}

					metb := tc.cs.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
					testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), metb)
					if metb.ForceRemoveSourceBrbnch != tc.removeSourceBrbnch {
						t.Fbtblf("unexpected ForceRemoveSourceBrbnch vblue: hbve %v, wbnt %v", metb.ForceRemoveSourceBrbnch, tc.removeSourceBrbnch)
					}
				})
			}
		})
	})

	t.Run("CloseChbngeset", func(t *testing.T) {
		t.Run("invblid metbdbtb", func(t *testing.T) {
			defer func() { _ = recover() }()

			p := newGitLbbChbngesetSourceTestProvider(t)
			repo := &types.Repo{Metbdbtb: struct{}{}}
			_ = p.source.CloseChbngeset(p.ctx, &Chbngeset{
				RemoteRepo: repo,
				TbrgetRepo: repo,
			})
			t.Error("invblid metbdbtb did not pbnic")
		})

		t.Run("error from UpdbteMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")
			mr := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockUpdbteMergeRequest(mr, nil, gitlbb.UpdbteMergeRequestStbteEventClose, inner)

			hbve := p.source.CloseChbngeset(p.ctx, p.chbngeset)
			if !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
		})

		t.Run("success", func(t *testing.T) {
			wbnt := &gitlbb.MergeRequest{}
			mr := &gitlbb.MergeRequest{IID: 2}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockUpdbteMergeRequest(mr, wbnt, gitlbb.UpdbteMergeRequestStbteEventClose, nil)
			p.mockGetMergeRequestNotes(mr.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(mr.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(mr.IID, nil, 20, nil)

			if err := p.source.CloseChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
		})
	})

	t.Run("ReopenChbngeset", func(t *testing.T) {
		t.Run("invblid metbdbtb", func(t *testing.T) {
			defer func() { _ = recover() }()

			p := newGitLbbChbngesetSourceTestProvider(t)
			repo := &types.Repo{Metbdbtb: struct{}{}}
			_ = p.source.ReopenChbngeset(p.ctx, &Chbngeset{
				RemoteRepo: repo,
				TbrgetRepo: repo,
			})
			t.Error("invblid metbdbtb did not pbnic")
		})

		t.Run("error from UpdbteMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")
			mr := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockUpdbteMergeRequest(mr, nil, gitlbb.UpdbteMergeRequestStbteEventReopen, inner)

			hbve := p.source.ReopenChbngeset(p.ctx, p.chbngeset)
			if !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
		})

		t.Run("success", func(t *testing.T) {
			wbnt := &gitlbb.MergeRequest{}
			mr := &gitlbb.MergeRequest{IID: 2}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockUpdbteMergeRequest(mr, wbnt, gitlbb.UpdbteMergeRequestStbteEventReopen, nil)
			p.mockGetMergeRequestNotes(mr.IID, nil, 20, nil)
			// TODO: bdd event
			p.mockGetMergeRequestResourceStbteEvents(mr.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(mr.IID, nil, 20, nil)

			if err := p.source.ReopenChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
		})
	})

	t.Run("LobdChbngeset", func(t *testing.T) {
		t.Run("invblid metbdbtb", func(t *testing.T) {
			defer func() { _ = recover() }()

			p := newGitLbbChbngesetSourceTestProvider(t)

			repo := &types.Repo{Metbdbtb: struct{}{}}
			_ = p.source.LobdChbngeset(p.ctx, &Chbngeset{
				RemoteRepo: repo,
				TbrgetRepo: repo,
			})
			t.Error("invblid metbdbtb did not pbnic")
		})

		t.Run("error from PbrseInt", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)
			repo := &types.Repo{Metbdbtb: &gitlbb.Project{}}
			if err := p.source.LobdChbngeset(p.ctx, &Chbngeset{
				Chbngeset: &btypes.Chbngeset{
					ExternblID: "foo",
					Metbdbtb:   &gitlbb.MergeRequest{},
				},
				RemoteRepo: repo,
				TbrgetRepo: repo,
			}); err == nil {
				t.Error("invblid ExternblID did not result in bn error")
			}
		})

		t.Run("error from GetMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, nil, inner)
			p.mockGetMergeRequestNotes(42, nil, 20, nil)
			p.mockGetMergeRequestPipelines(42, nil, 20, nil)

			if hbve := p.source.LobdChbngeset(p.ctx, p.chbngeset); !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
		})

		t.Run("error from GetMergeRequestNotes", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}
			inner := errors.New("foo")

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, nil, 20, inner)
			p.mockGetMergeRequestResourceStbteEvents(43, nil, 20, nil)
			p.mockGetMergeRequestPipelines(43, nil, 20, nil)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); !errors.Is(err, inner) {
				t.Errorf("unexpected error: %+v", err)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != p.mr {
				t.Errorf("metbdbtb unexpectedly chbnged to %+v", p.chbngeset.Chbngeset.Metbdbtb)
			}
		})

		t.Run("error from GetMergeRequestResourceStbteEvents", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}
			inner := errors.New("foo")

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(43, nil, 20, inner)
			p.mockGetMergeRequestPipelines(43, nil, 20, nil)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); !errors.Is(err, inner) {
				t.Errorf("unexpected error: %+v", err)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != p.mr {
				t.Errorf("metbdbtb unexpectedly chbnged to %+v", p.chbngeset.Chbngeset.Metbdbtb)
			}
		})

		t.Run("error from GetMergeRequestPipelines", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}
			inner := errors.New("foo")

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(43, nil, 20, nil)
			p.mockGetMergeRequestPipelines(43, nil, 20, inner)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); !errors.Is(err, inner) {
				t.Errorf("unexpected error: %+v", err)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != p.mr {
				t.Errorf("metbdbtb unexpectedly chbnged to %+v", p.chbngeset.Chbngeset.Metbdbtb)
			}
		})

		t.Run("success", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(43, nil, 20, nil)
			p.mockGetMergeRequestPipelines(43, nil, 20, nil)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve := p.chbngeset.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest); hbve != mr {
				t.Errorf("merge request metbdbtb not updbted: hbve %p; wbnt %p", hbve, mr)
			}
		})

		t.Run("not found", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "43"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(43, nil, gitlbb.ErrMergeRequestNotFound)

			expected := ChbngesetNotFoundError{
				Chbngeset: &Chbngeset{
					Chbngeset: &btypes.Chbngeset{ExternblID: "43"},
				},
			}

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err == nil {
				t.Fbtbl("unexpectedly no error for not found chbngeset")
			} else if !errors.Is(err, expected) {
				t.Fbtblf("unexpected error: %+v", err)
			}
		})

		t.Run("integrbtion", func(t *testing.T) {
			repo := &types.Repo{
				Metbdbtb: &gitlbb.Project{
					// sourcegrbph/sourcegrbph
					ProjectCommon: gitlbb.ProjectCommon{ID: 16606088},
				},
			}
			testCbses := []struct {
				nbme string
				cs   *Chbngeset
				err  string
			}{
				{
					nbme: "found",
					cs: &Chbngeset{
						RemoteRepo: repo,
						TbrgetRepo: repo,
						Chbngeset:  &btypes.Chbngeset{ExternblID: "2"},
					},
				},
				{
					nbme: "not-found",
					cs: &Chbngeset{
						RemoteRepo: repo,
						TbrgetRepo: repo,
						Chbngeset:  &btypes.Chbngeset{ExternblID: "100000"},
					},
					err: "Chbngeset with externbl ID 100000 not found",
				},
				{
					nbme: "project-not-found",
					cs: &Chbngeset{
						RemoteRepo: repo,
						TbrgetRepo: &types.Repo{Metbdbtb: &gitlbb.Project{
							ProjectCommon: gitlbb.ProjectCommon{ID: 999999999999},
						}},
						Chbngeset: &btypes.Chbngeset{ExternblID: "100000"},
					},
					// Not b chbngeset not found error. This is importbnt so we don't set
					// b chbngeset bs deleted, when the token scope cbnnot view the project
					// the MR lives in.
					err: "retrieving merge request 100000: sending request to get b merge request: GitLbb project not found",
				},
			}

			for _, tc := rbnge testCbses {
				tc := tc
				tc.nbme = "GitlbbSource_LobdChbngeset_" + tc.nbme

				t.Run(tc.nbme, func(t *testing.T) {
					cf, sbve := newClientFbctory(t, tc.nbme)
					defer sbve(t)

					lg := log15.New()
					lg.SetHbndler(log15.DiscbrdHbndler())

					svc := &types.ExternblService{
						Kind: extsvc.KindGitLbb,
						Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitLbbConnection{
							Url:   "https://gitlbb.com",
							Token: os.Getenv("GITLAB_TOKEN"),
						})),
					}

					ctx := context.Bbckground()
					gitlbbSource, err := NewGitLbbSource(ctx, svc, cf)
					if err != nil {
						t.Fbtbl(err)
					}

					if tc.err == "" {
						tc.err = "<nil>"
					}

					err = gitlbbSource.LobdChbngeset(ctx, tc.cs)
					if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
						t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
					}

					if err != nil {
						return
					}

					metb := tc.cs.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
					testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), metb)
				})
			}
		})

		// The guts of the note bnd pipeline scenbrios bre tested in sepbrbte
		// unit tests below for rebd{Notes,Pipelines}UntilSeen, but we'll do b
		// couple of quick tests here just to ensure thbt
		// decorbteMergeRequestDbtb does the right thing.
		t.Run("notes", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}
			notes := []*gitlbb.Note{
				{ID: 1, System: true},
				{ID: 2, System: true},
				{ID: 3, System: fblse},
			}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, notes, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(43, nil, 20, nil)
			p.mockGetMergeRequestPipelines(43, nil, 20, nil)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if diff := cmp.Diff(mr.Notes, notes[0:2]); diff != "" {
				t.Errorf("unexpected notes: %s", diff)
			}

			// A subsequent lobd should result in the sbme notes. Since we
			// chbnged the IID in the merge request, we do need to chbnge the
			// getMergeRequest mock.
			p.mockGetMergeRequest(43, mr, nil)
			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if diff := cmp.Diff(mr.Notes, notes[0:2]); diff != "" {
				t.Errorf("unexpected notes: %s", diff)
			}
		})

		t.Run("resource stbte events", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}
			events := []*gitlbb.ResourceStbteEvent{
				{
					ID:    1,
					Stbte: gitlbb.ResourceStbteEventStbteClosed,
				},
				{
					ID:    2,
					Stbte: gitlbb.ResourceStbteEventStbteMerged,
				},
				{
					ID:    3,
					Stbte: gitlbb.ResourceStbteEventStbteReopened,
				},
			}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(43, events, 20, nil)
			p.mockGetMergeRequestPipelines(43, nil, 20, nil)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if diff := cmp.Diff(mr.ResourceStbteEvents, events); diff != "" {
				t.Errorf("unexpected events: %s", diff)
			}

			// A subsequent lobd should result in the sbme events. Since we
			// chbnged the IID in the merge request, we do need to chbnge the
			// getMergeRequest mock.
			p.mockGetMergeRequest(43, mr, nil)
			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if diff := cmp.Diff(mr.ResourceStbteEvents, events); diff != "" {
				t.Errorf("unexpected events: %s", diff)
			}
		})

		t.Run("pipelines", func(t *testing.T) {
			// A new merge request with b new IID.
			mr := &gitlbb.MergeRequest{IID: 43}
			pipelines := []*gitlbb.Pipeline{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.ExternblID = "42"
			p.chbngeset.Chbngeset.Metbdbtb = p.mr
			p.mockGetMergeRequest(42, mr, nil)
			p.mockGetMergeRequestNotes(43, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(43, nil, 20, nil)
			p.mockGetMergeRequestPipelines(43, pipelines, 20, nil)

			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if diff := cmp.Diff(mr.Pipelines, pipelines); diff != "" {
				t.Errorf("unexpected pipelines: %s", diff)
			}

			// A subsequent lobd should result in the sbme pipelines. Since we
			// chbnged the IID in the merge request, we do need to chbnge the
			// getMergeRequest mock.
			p.mockGetMergeRequest(43, mr, nil)
			if err := p.source.LobdChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if diff := cmp.Diff(mr.Pipelines, pipelines); diff != "" {
				t.Errorf("unexpected pipelines: %s", diff)
			}
		})
	})

	t.Run("UpdbteChbngeset", func(t *testing.T) {
		t.Run("invblid metbdbtb", func(t *testing.T) {
			p := newGitLbbChbngesetSourceTestProvider(t)

			err := p.source.UpdbteChbngeset(p.ctx, &Chbngeset{
				Chbngeset: &btypes.Chbngeset{Metbdbtb: struct{}{}},
			})
			if err == nil {
				t.Error("unexpected nil error")
			}
		})

		t.Run("error from UpdbteMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")
			mr := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockUpdbteMergeRequest(mr, nil, "", inner)

			hbve := p.source.UpdbteChbngeset(p.ctx, p.chbngeset)
			if !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != mr {
				t.Errorf("metbdbtb unexpectedly updbted: from %+v; to %+v", mr, p.chbngeset.Chbngeset.Metbdbtb)
			}
		})

		t.Run("success", func(t *testing.T) {
			in := &gitlbb.MergeRequest{IID: 2}
			out := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = in
			p.mockUpdbteMergeRequest(in, out, "", nil)
			p.mockGetMergeRequestNotes(in.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(in.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(in.IID, nil, 20, nil)

			if err := p.source.UpdbteChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected non-nil error: %+v", err)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != out {
				t.Errorf("metbdbtb not correctly updbted: hbve %+v; wbnt %+v", p.chbngeset.Chbngeset.Metbdbtb, out)
			}
		})
	})

	t.Run("UpdbteChbngeset drbft", func(t *testing.T) {
		t.Run("GitLbb version is grebter thbn 14.0.0", func(t *testing.T) {
			// We won't test the full set of UpdbteChbngeset scenbrios; instebd
			// we'll just mbke sure the title is bppropribtely munged.
			in := &gitlbb.MergeRequest{IID: 2, WorkInProgress: true}
			out := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.mockGetVersions(mockVersion2.String(), p.source.client.Urn())
			p.chbngeset.Chbngeset.Metbdbtb = in

			oldMock := gitlbb.MockUpdbteMergeRequest
			t.Clebnup(func() { gitlbb.MockUpdbteMergeRequest = oldMock })
			gitlbb.MockUpdbteMergeRequest = func(c *gitlbb.Client, ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest, opts gitlbb.UpdbteMergeRequestOpts) (*gitlbb.MergeRequest, error) {
				if hbve, wbnt := opts.Title, "Drbft: title"; hbve != wbnt {
					t.Errorf("unexpected title: hbve=%q wbnt=%q", hbve, wbnt)
				}
				return out, nil
			}

			p.mockGetMergeRequestNotes(in.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(in.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(in.IID, nil, 20, nil)

			if err := p.source.UpdbteChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected non-nil error: %+v", err)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != out {
				t.Errorf("metbdbtb not correctly updbted: hbve %+v; wbnt %+v", p.chbngeset.Chbngeset.Metbdbtb, out)
			}
		})

		t.Run("GitLbb version is less thbn 14.0.0", func(t *testing.T) {
			// We won't test the full set of UpdbteChbngeset scenbrios; instebd
			// we'll just mbke sure the title is bppropribtely munged.
			in := &gitlbb.MergeRequest{IID: 2, WorkInProgress: true}
			out := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.mockGetVersions(mockVersion.String(), p.source.client.Urn())
			p.chbngeset.Chbngeset.Metbdbtb = in

			oldMock := gitlbb.MockUpdbteMergeRequest
			t.Clebnup(func() { gitlbb.MockUpdbteMergeRequest = oldMock })
			gitlbb.MockUpdbteMergeRequest = func(c *gitlbb.Client, ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest, opts gitlbb.UpdbteMergeRequestOpts) (*gitlbb.MergeRequest, error) {
				if hbve, wbnt := opts.Title, "WIP: title"; hbve != wbnt {
					t.Errorf("unexpected title: hbve=%q wbnt=%q", hbve, wbnt)
				}
				return out, nil
			}

			p.mockGetMergeRequestNotes(in.IID, nil, 20, nil)
			p.mockGetMergeRequestResourceStbteEvents(in.IID, nil, 20, nil)
			p.mockGetMergeRequestPipelines(in.IID, nil, 20, nil)

			if err := p.source.UpdbteChbngeset(p.ctx, p.chbngeset); err != nil {
				t.Errorf("unexpected non-nil error: %+v", err)
			}
			if p.chbngeset.Chbngeset.Metbdbtb != out {
				t.Errorf("metbdbtb not correctly updbted: hbve %+v; wbnt %+v", p.chbngeset.Chbngeset.Metbdbtb, out)
			}
		})
	})

	t.Run("UndrbftChbngeset", func(t *testing.T) {
		in := &gitlbb.MergeRequest{IID: 2, WorkInProgress: true}
		out := &gitlbb.MergeRequest{}

		p := newGitLbbChbngesetSourceTestProvider(t)
		p.chbngeset.Chbngeset.Metbdbtb = in

		oldMock := gitlbb.MockUpdbteMergeRequest
		t.Clebnup(func() { gitlbb.MockUpdbteMergeRequest = oldMock })
		gitlbb.MockUpdbteMergeRequest = func(c *gitlbb.Client, ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest, opts gitlbb.UpdbteMergeRequestOpts) (*gitlbb.MergeRequest, error) {
			if hbve, wbnt := opts.Title, "title"; hbve != wbnt {
				t.Errorf("unexpected title: hbve=%q wbnt=%q", hbve, wbnt)
			}
			return out, nil
		}

		p.mockGetVersions(mockVersion.String(), p.source.client.Urn())
		p.mockGetMergeRequestNotes(in.IID, nil, 20, nil)
		p.mockGetMergeRequestResourceStbteEvents(in.IID, nil, 20, nil)
		p.mockGetMergeRequestPipelines(in.IID, nil, 20, nil)

		if err := p.source.UndrbftChbngeset(p.ctx, p.chbngeset); err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		if p.chbngeset.Chbngeset.Metbdbtb != out {
			t.Errorf("metbdbtb not correctly updbted: hbve %+v; wbnt %+v", p.chbngeset.Chbngeset.Metbdbtb, out)
		}
	})

	t.Run("CrebteComment", func(t *testing.T) {
		commentBody := "test-comment"
		t.Run("invblid metbdbtb", func(t *testing.T) {
			defer func() { _ = recover() }()

			p := newGitLbbChbngesetSourceTestProvider(t)
			repo := &types.Repo{Metbdbtb: struct{}{}}
			_ = p.source.CrebteComment(p.ctx, &Chbngeset{
				RemoteRepo: repo,
				TbrgetRepo: repo,
			}, commentBody)
			t.Error("invblid metbdbtb did not pbnic")
		})

		t.Run("error from CrebteComment", func(t *testing.T) {
			inner := errors.New("foo")
			mr := &gitlbb.MergeRequest{}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockCrebteComment(commentBody, inner)

			hbve := p.source.CrebteComment(p.ctx, p.chbngeset, commentBody)
			if !errors.Is(hbve, inner) {
				t.Errorf("error does not include inner error: hbve %+v; wbnt %+v", hbve, inner)
			}
		})

		t.Run("success", func(t *testing.T) {
			mr := &gitlbb.MergeRequest{IID: 2}

			p := newGitLbbChbngesetSourceTestProvider(t)
			p.chbngeset.Chbngeset.Metbdbtb = mr
			p.mockCrebteComment(commentBody, nil)

			if err := p.source.CrebteComment(p.ctx, p.chbngeset, commentBody); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
		})

		t.Run("integrbtion", func(t *testing.T) {
			nbme := "GitlbbSource_CrebteComment_success"

			t.Run(nbme, func(t *testing.T) {
				cf, sbve := newClientFbctory(t, nbme)
				defer sbve(t)

				lg := log15.New()
				lg.SetHbndler(log15.DiscbrdHbndler())

				svc := &types.ExternblService{
					Kind: extsvc.KindGitLbb,
					Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitLbbConnection{
						Url:   "https://gitlbb.com",
						Token: os.Getenv("GITLAB_TOKEN"),
					})),
				}

				ctx := context.Bbckground()
				gitlbbSource, err := NewGitLbbSource(ctx, svc, cf)
				if err != nil {
					t.Fbtbl(err)
				}

				repo := &types.Repo{Metbdbtb: newGitLbbProject(16606088)}
				cs := &Chbngeset{
					RemoteRepo: repo,
					TbrgetRepo: repo,
					Chbngeset: &btypes.Chbngeset{Metbdbtb: &gitlbb.MergeRequest{
						IID: gitlbb.ID(2),
					}},
				}

				if err := gitlbbSource.CrebteComment(ctx, cs, "test-comment"); err != nil {
					t.Fbtbl(err)
				}
			})
		})
	})
}

func TestRebdNotesUntilSeen(t *testing.T) {
	commonNotes := []*gitlbb.Note{
		{ID: 1, System: true},
		{ID: 2, System: true},
		{ID: 3, System: true},
		{ID: 4, System: true},
	}

	t.Run("rebds bll notes", func(t *testing.T) {
		notes, err := rebdSystemNotes(pbginbtedNoteIterbtor(commonNotes, 2))
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		if diff := cmp.Diff(notes, commonNotes); diff != "" {
			t.Errorf("unexpected notes: %s", diff)
		}
	})

	t.Run("error from iterbtor", func(t *testing.T) {
		wbnt := errors.New("foo")
		notes, err := rebdSystemNotes(func() ([]*gitlbb.Note, error) {
			return nil, wbnt
		})
		if notes != nil {
			t.Errorf("unexpected non-nil notes: %+v", notes)
		}
		if !errors.Is(err, wbnt) {
			t.Errorf("expected error not found in chbin: hbve %+v; wbnt %+v", err, wbnt)
		}
	})

	t.Run("no system notes", func(t *testing.T) {
		notes, err := rebdSystemNotes(pbginbtedNoteIterbtor([]*gitlbb.Note{
			{ID: 1, System: fblse},
			{ID: 2, System: fblse},
			{ID: 3, System: fblse},
			{ID: 4, System: fblse},
		}, 2))
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		if len(notes) > 0 {
			t.Errorf("unexpected notes: %+v", notes)
		}
	})

	t.Run("no pbges", func(t *testing.T) {
		notes, err := rebdSystemNotes(pbginbtedNoteIterbtor([]*gitlbb.Note{}, 2))
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		if len(notes) > 0 {
			t.Errorf("unexpected notes: %+v", notes)
		}
	})
}

func TestRebdPipelinesUntilSeen(t *testing.T) {
	commonPipelines := []*gitlbb.Pipeline{
		{ID: 1},
		{ID: 2},
		{ID: 3},
		{ID: 4},
	}

	t.Run("rebds bll pipelines", func(t *testing.T) {
		notes, err := rebdPipelines(pbginbtedPipelineIterbtor(commonPipelines, 2))
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		if diff := cmp.Diff(notes, commonPipelines); diff != "" {
			t.Errorf("unexpected notes: %s", diff)
		}
	})

	t.Run("error from iterbtor", func(t *testing.T) {
		wbnt := errors.New("foo")
		pipelines, err := rebdPipelines(func() ([]*gitlbb.Pipeline, error) {
			return nil, wbnt
		})
		if pipelines != nil {
			t.Errorf("unexpected non-nil pipelines: %+v", pipelines)
		}
		if !errors.Is(err, wbnt) {
			t.Errorf("expected error not found in chbin: hbve %+v; wbnt %+v", err, wbnt)
		}
	})

	t.Run("no pbges", func(t *testing.T) {
		pipelines, err := rebdPipelines(pbginbtedPipelineIterbtor([]*gitlbb.Pipeline{}, 2))
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		if len(pipelines) > 0 {
			t.Errorf("unexpected pipelines: %+v", pipelines)
		}
	})
}

type gitLbbChbngesetSourceTestProvider struct {
	chbngeset *Chbngeset
	ctx       context.Context
	mr        *gitlbb.MergeRequest
	source    *GitLbbSource
	t         *testing.T

	isGetVersionCblled bool
}

// newGitLbbChbngesetSourceTestProvider provides b set of useful pre-cbnned
// objects, blong with b hbndful of methods to mock underlying
// internbl/extsvc/gitlbb functions.
func newGitLbbChbngesetSourceTestProvider(t *testing.T) *gitLbbChbngesetSourceTestProvider {
	prov := gitlbb.NewClientProvider("Test", &url.URL{}, &pbnicDoer{})
	repo := &types.Repo{Metbdbtb: &gitlbb.Project{}}
	p := &gitLbbChbngesetSourceTestProvider{
		chbngeset: &Chbngeset{
			Chbngeset:  &btypes.Chbngeset{},
			RemoteRepo: repo,
			TbrgetRepo: repo,
			HebdRef:    "refs/hebds/hebd",
			BbseRef:    "refs/hebds/bbse",
			Title:      "title",
			Body:       "description",
		},
		ctx: context.Bbckground(),
		mr: &gitlbb.MergeRequest{
			ID:              1,
			IID:             2,
			ProjectID:       3,
			SourceProjectID: 3,
			Title:           "title",
			Description:     "description",
			SourceBrbnch:    "hebd",
			TbrgetBrbnch:    "bbse",
		},
		source: &GitLbbSource{
			client: prov.GetClient(),
		},
		t: t,
	}

	// Rbther thbn requiring the cbller to defer b cbll to unmock, we cbn do it
	// here bnd be sure we'll hbve it done when the test is complete.
	t.Clebnup(func() { p.unmock() })

	return p
}

func (p *gitLbbChbngesetSourceTestProvider) testCommonPbrbms(ctx context.Context, client *gitlbb.Client, project *gitlbb.Project) {
	if client != p.source.client {
		p.t.Errorf("unexpected GitLbbSource client: hbve %+v; wbnt %+v", client, p.source.client)
	}
	if ctx != p.ctx {
		p.t.Errorf("unexpected context: hbve %+v; wbnt %+v", ctx, p.ctx)
	}
	if project != p.chbngeset.TbrgetRepo.Metbdbtb.(*gitlbb.Project) {
		p.t.Errorf("unexpected Project: hbve %+v; wbnt %+v", project, p.chbngeset.TbrgetRepo.Metbdbtb)
	}
}

// mockCrebteMergeRequest mocks b gitlbb.CrebteMergeRequest cbll. Note thbt only
// the SourceBrbnch bnd TbrgetBrbnch fields of the expected options bre checked.
func (p *gitLbbChbngesetSourceTestProvider) mockCrebteMergeRequest(expected gitlbb.CrebteMergeRequestOpts, mr *gitlbb.MergeRequest, err error) {
	gitlbb.MockCrebteMergeRequest = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, opts gitlbb.CrebteMergeRequestOpts) (*gitlbb.MergeRequest, error) {
		p.testCommonPbrbms(ctx, client, project)

		if wbnt := expected.SourceBrbnch; opts.SourceBrbnch != wbnt {
			p.t.Errorf("unexpected SourceBrbnch: hbve %s; wbnt %s", opts.SourceBrbnch, wbnt)
		}
		if wbnt := expected.TbrgetBrbnch; opts.TbrgetBrbnch != wbnt {
			p.t.Errorf("unexpected TbrgetBrbnch: hbve %s; wbnt %s", opts.TbrgetBrbnch, wbnt)
		}

		return mr, err
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetMergeRequest(expected gitlbb.ID, mr *gitlbb.MergeRequest, err error) {
	gitlbb.MockGetMergeRequest = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, iid gitlbb.ID) (*gitlbb.MergeRequest, error) {
		p.testCommonPbrbms(ctx, client, project)
		if expected != iid {
			p.t.Errorf("invblid IID: hbve %d; wbnt %d", iid, expected)
		}
		return mr, err
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetMergeRequestNotes(expectedIID gitlbb.ID, notes []*gitlbb.Note, pbgeSize int, err error) {
	gitlbb.MockGetMergeRequestNotes = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, iid gitlbb.ID) func() ([]*gitlbb.Note, error) {
		p.testCommonPbrbms(ctx, client, project)
		if expectedIID != iid {
			p.t.Errorf("unexpected IID: hbve %d; wbnt %d", iid, expectedIID)
		}

		if err != nil {
			return func() ([]*gitlbb.Note, error) { return nil, err }
		}
		return pbginbtedNoteIterbtor(notes, pbgeSize)
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetMergeRequestResourceStbteEvents(expectedIID gitlbb.ID, events []*gitlbb.ResourceStbteEvent, pbgeSize int, err error) {
	gitlbb.MockGetMergeRequestResourceStbteEvents = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, iid gitlbb.ID) func() ([]*gitlbb.ResourceStbteEvent, error) {
		p.testCommonPbrbms(ctx, client, project)
		if expectedIID != iid {
			p.t.Errorf("unexpected IID: hbve %d; wbnt %d", iid, expectedIID)
		}

		if err != nil {
			return func() ([]*gitlbb.ResourceStbteEvent, error) { return nil, err }
		}
		return pbginbtedResourceStbteEventIterbtor(events, pbgeSize)
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetMergeRequestPipelines(expectedIID gitlbb.ID, pipelines []*gitlbb.Pipeline, pbgeSize int, err error) {
	gitlbb.MockGetMergeRequestPipelines = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, iid gitlbb.ID) func() ([]*gitlbb.Pipeline, error) {
		p.testCommonPbrbms(ctx, client, project)
		if expectedIID != iid {
			p.t.Errorf("unexpected IID: hbve %d; wbnt %d", iid, expectedIID)
		}

		if err != nil {
			return func() ([]*gitlbb.Pipeline, error) { return nil, err }
		}
		return pbginbtedPipelineIterbtor(pipelines, pbgeSize)
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetOpenMergeRequestByRefs(mr *gitlbb.MergeRequest, err error) {
	gitlbb.MockGetOpenMergeRequestByRefs = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, source, tbrget string) (*gitlbb.MergeRequest, error) {
		p.testCommonPbrbms(ctx, client, project)
		return mr, err
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockUpdbteMergeRequest(expectedMR, updbted *gitlbb.MergeRequest, expectedStbteEvent gitlbb.UpdbteMergeRequestStbteEvent, err error) {
	gitlbb.MockUpdbteMergeRequest = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, mrIn *gitlbb.MergeRequest, opts gitlbb.UpdbteMergeRequestOpts) (*gitlbb.MergeRequest, error) {
		p.testCommonPbrbms(ctx, client, project)
		if expectedMR != mrIn {
			p.t.Errorf("unexpected MergeRequest: hbve %+v; wbnt %+v", mrIn, expectedMR)
		}
		if len(expectedStbteEvent) != 0 && opts.StbteEvent != expectedStbteEvent {
			p.t.Errorf("unexpected StbteEvent: hbve %+v; wbnt %+v", opts.StbteEvent, expectedStbteEvent)
		}

		return updbted, err
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockCrebteComment(expected string, err error) {
	gitlbb.MockCrebteMergeRequestNote = func(client *gitlbb.Client, ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest, body string) error {
		p.testCommonPbrbms(ctx, client, project)
		if expected != body {
			p.t.Errorf("invblid body pbssed: hbve %q; wbnt %q", body, expected)
		}
		return err
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetVersions(expected, key string) {
	versions.MockGetVersions = func() ([]*versions.Version, error) {
		return []*versions.Version{
			{
				ExternblServiceKind: extsvc.KindGitLbb,
				Version:             expected,
				Key:                 key,
			},
			{
				ExternblServiceKind: extsvc.KindGitHub,
				Version:             "2.38.0",
				Key:                 "rbndom-key-<1>",
			},
			{
				ExternblServiceKind: extsvc.KindGitLbb,
				Version:             "1.3.5",
				Key:                 "rbndom-key-<2>",
			},
			{
				ExternblServiceKind: extsvc.KindBitbucketCloud,
				Version:             "1.2.5",
				Key:                 "rbndom-key-<3>",
			},
		}, nil
	}
}

func (p *gitLbbChbngesetSourceTestProvider) mockGetVersion(expected string) {
	gitlbb.MockGetVersion = func(ctx context.Context) (string, error) {
		p.isGetVersionCblled = true
		return expected, nil
	}
}

func (p *gitLbbChbngesetSourceTestProvider) unmock() {
	gitlbb.MockCrebteMergeRequest = nil
	gitlbb.MockGetMergeRequest = nil
	gitlbb.MockGetMergeRequestNotes = nil
	gitlbb.MockGetMergeRequestResourceStbteEvents = nil
	gitlbb.MockGetMergeRequestPipelines = nil
	gitlbb.MockGetOpenMergeRequestByRefs = nil
	gitlbb.MockUpdbteMergeRequest = nil
	gitlbb.MockCrebteMergeRequestNote = nil

	versions.MockGetVersions = nil
}

// pbnicDoer provides b httpcli.Doer implementbtion thbt pbnics if bny bttempt
// is mbde to issue b HTTP request; thereby ensuring thbt our unit tests don't
// bctublly try to tblk to GitLbb.
type pbnicDoer struct{}

func (d *pbnicDoer) Do(r *http.Request) (*http.Response, error) {
	pbnic("this function should not be cblled; b mock must be missing")
}

// pbginbtedNoteIterbtor essentiblly fbkes the pbginbtion behbviour implemented
// by gitlbb.GetMergeRequestNotes with b cbnned notes list.
func pbginbtedNoteIterbtor(notes []*gitlbb.Note, pbgeSize int) func() ([]*gitlbb.Note, error) {
	pbge := 0

	return func() ([]*gitlbb.Note, error) {
		low := pbgeSize * pbge
		high := pbgeSize * (pbge + 1)
		pbge++

		if low >= len(notes) {
			return []*gitlbb.Note{}, nil
		}
		if high > len(notes) {
			return notes[low:], nil
		}
		return notes[low:high], nil
	}
}

// pbginbtedResourceStbteEventIterbtor essentiblly fbkes the pbginbtion behbviour implemented
// by gitlbb.GetMergeRequestResourceStbteEvents with b cbnned resource stbte events list.
func pbginbtedResourceStbteEventIterbtor(events []*gitlbb.ResourceStbteEvent, pbgeSize int) func() ([]*gitlbb.ResourceStbteEvent, error) {
	pbge := 0

	return func() ([]*gitlbb.ResourceStbteEvent, error) {
		low := pbgeSize * pbge
		high := pbgeSize * (pbge + 1)
		pbge++

		if low >= len(events) {
			return []*gitlbb.ResourceStbteEvent{}, nil
		}
		if high > len(events) {
			return events[low:], nil
		}
		return events[low:high], nil
	}
}

// pbginbtedPipelineIterbtor essentiblly fbkes the pbginbtion behbviour
// implemented by gitlbb.GetMergeRequestPipelines with b cbnned pipelines list.
func pbginbtedPipelineIterbtor(pipelines []*gitlbb.Pipeline, pbgeSize int) func() ([]*gitlbb.Pipeline, error) {
	pbge := 0

	return func() ([]*gitlbb.Pipeline, error) {
		low := pbgeSize * pbge
		high := pbgeSize * (pbge + 1)
		pbge++

		if low >= len(pipelines) {
			return []*gitlbb.Pipeline{}, nil
		}
		if high > len(pipelines) {
			return pipelines[low:], nil
		}
		return pipelines[low:high], nil
	}
}

func TestGitLbbSource_WithAuthenticbtor(t *testing.T) {
	t.Run("supported", func(t *testing.T) {
		vbr src ChbngesetSource
		src, err := newGitLbbSource("Test", &schemb.GitLbbConnection{}, nil)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		src, err = src.WithAuthenticbtor(&buth.OAuthBebrerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GitLbbSource); !ok {
			t.Error("cbnnot coerce Source into GitLbbSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]buth.Authenticbtor{
			"nil":         nil,
			"BbsicAuth":   &buth.BbsicAuth{},
			"OAuthClient": &buth.OAuthClient{},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr src ChbngesetSource
				src, err := newGitLbbSource("Test", &schemb.GitLbbConnection{}, nil)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				src, err = src.WithAuthenticbtor(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HbsType(err, UnsupportedAuthenticbtorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestGitlbbSource_GetFork(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("fbilures", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			tbrgetRepo *types.Repo
			client     gitlbbClientFork
		}{
			"invblid PbthWithNbmespbce": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &gitlbb.Project{
						ProjectCommon: gitlbb.ProjectCommon{
							PbthWithNbmespbce: "foo",
						},
					},
				},
				client: nil,
			},
			"client error": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &gitlbb.Project{
						ProjectCommon: gitlbb.ProjectCommon{
							PbthWithNbmespbce: "foo/bbr",
						},
					},
				},
				client: &mockGitlbbClientFork{err: errors.New("hello!")},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				fork, err := getGitLbbForkInternbl(ctx, tc.tbrgetRepo, tc.client, nil, nil)
				bssert.Nil(t, fork)
				bssert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		org := "org"
		user := "user"
		urn := extsvc.URN(extsvc.KindGitLbb, 1)

		for nbme, tc := rbnge mbp[string]struct {
			tbrgetRepo    *types.Repo
			forkRepo      *gitlbb.Project
			nbmespbce     *string
			wbntNbmespbce string
			nbme          *string
			wbntNbme      string
			client        gitlbbClientFork
		}{
			"no nbmespbce": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &gitlbb.Project{
						ProjectCommon: gitlbb.ProjectCommon{
							ID:                1,
							PbthWithNbmespbce: "foo/bbr",
						},
					},
					Sources: mbp[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://gitlbb.com/foo/bbr",
						},
					},
				},
				forkRepo: &gitlbb.Project{
					ForkedFromProject: &gitlbb.ProjectCommon{ID: 1},
					ProjectCommon:     gitlbb.ProjectCommon{ID: 2, PbthWithNbmespbce: user + "/user-bbr"}},
				nbmespbce:     nil,
				wbntNbmespbce: user,
				wbntNbme:      user + "-bbr",
				client: &mockGitlbbClientFork{fork: &gitlbb.Project{ForkedFromProject: &gitlbb.ProjectCommon{ID: 1},
					ProjectCommon: gitlbb.ProjectCommon{ID: 2, PbthWithNbmespbce: user + "/user-bbr"}}},
			},
			"with nbmespbce": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &gitlbb.Project{
						ProjectCommon: gitlbb.ProjectCommon{
							ID:                1,
							PbthWithNbmespbce: "foo/bbr",
						},
					},
					Sources: mbp[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://gitlbb.com/foo/bbr",
						},
					},
				},
				forkRepo: &gitlbb.Project{
					ForkedFromProject: &gitlbb.ProjectCommon{ID: 1},
					ProjectCommon:     gitlbb.ProjectCommon{ID: 2, PbthWithNbmespbce: org + "/" + org + "-bbr"}},
				nbmespbce:     &org,
				wbntNbmespbce: org,
				wbntNbme:      org + "-bbr",
				client: &mockGitlbbClientFork{
					fork: &gitlbb.Project{
						ForkedFromProject: &gitlbb.ProjectCommon{ID: 1},
						ProjectCommon:     gitlbb.ProjectCommon{ID: 2, PbthWithNbmespbce: org + "/" + org + "-bbr"}},
					wbntOrg: &org,
				},
			},
			"with nbmespbce bnd nbme": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &gitlbb.Project{
						ProjectCommon: gitlbb.ProjectCommon{
							ID:                1,
							PbthWithNbmespbce: "foo/bbr",
						},
					},
					Sources: mbp[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://gitlbb.com/foo/bbr",
						},
					},
				},
				forkRepo: &gitlbb.Project{
					ForkedFromProject: &gitlbb.ProjectCommon{ID: 1},
					ProjectCommon:     gitlbb.ProjectCommon{ID: 2, PbthWithNbmespbce: org + "/custom-bbr"}},
				nbmespbce:     &org,
				wbntNbmespbce: org,
				nbme:          pointers.Ptr("custom-bbr"),
				wbntNbme:      "custom-bbr",
				client: &mockGitlbbClientFork{
					fork: &gitlbb.Project{
						ForkedFromProject: &gitlbb.ProjectCommon{ID: 1},
						ProjectCommon:     gitlbb.ProjectCommon{ID: 2, PbthWithNbmespbce: org + "/custom-bbr"}},
					wbntOrg: &org,
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				fork, err := getGitLbbForkInternbl(ctx, tc.tbrgetRepo, tc.client, tc.nbmespbce, tc.nbme)
				bssert.Nil(t, err)
				bssert.NotNil(t, fork)
				bssert.NotEqubl(t, fork, tc.tbrgetRepo)
				bssert.Equbl(t, tc.forkRepo, fork.Metbdbtb)
				bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://gitlbb.com/"+tc.wbntNbmespbce+"/"+tc.wbntNbme)
			})
		}
	})
}

type mockGitlbbClientFork struct {
	wbntOrg *string
	fork    *gitlbb.Project
	err     error
}

vbr _ gitlbbClientFork = &mockGitlbbClientFork{}

func (mock *mockGitlbbClientFork) ForkProject(ctx context.Context, project *gitlbb.Project, nbmespbce *string, nbme string) (*gitlbb.Project, error) {
	if (mock.wbntOrg == nil && nbmespbce != nil) || (mock.wbntOrg != nil && nbmespbce == nil) || (mock.wbntOrg != nil && nbmespbce != nil && *mock.wbntOrg != *nbmespbce) {
		return nil, errors.Newf("unexpected orgbnisbtion: hbve=%v wbnt=%v", nbmespbce, mock.wbntOrg)
	}

	return mock.fork, mock.err
}

func TestDecorbteMergeRequestDbtb(t *testing.T) {
	ctx := context.Bbckground()

	// The test fixtures use publicly bvbilbble merge requests, bnd should be
	// bble to be updbted bt bny time without bny bction required.
	crebteSource := func(t *testing.T) *GitLbbSource {
		cf, sbve := newClientFbctory(t, t.Nbme())
		t.Clebnup(func() { sbve(t) })

		src, err := newGitLbbSource(
			"Test",
			&schemb.GitLbbConnection{
				Url:   "https://gitlbb.com",
				Token: os.Getenv("GITLAB_TOKEN"),
			},
			cf,
		)

		bssert.Nil(t, err)
		return src
	}

	src := crebteSource(t)

	// https://gitlbb.com/sourcegrbph/src-cli/-/merge_requests/6
	forked, err := src.client.GetMergeRequest(ctx, newGitLbbProject(16606399), 6)
	bssert.Nil(t, err)

	// https://gitlbb.com/sourcegrbph/sourcegrbph/-/merge_requests/1
	unforked, err := src.client.GetMergeRequest(ctx, newGitLbbProject(16606088), 1)
	bssert.Nil(t, err)

	t.Run("fork", func(t *testing.T) {
		err := crebteSource(t).decorbteMergeRequestDbtb(ctx, newGitLbbProject(int(forked.ProjectID)), forked)
		bssert.Nil(t, err)
		bssert.Equbl(t, "courier-new", forked.SourceProjectNbmespbce)
		bssert.Equbl(t, "src-cli-forked", forked.SourceProjectNbme)
	})

	t.Run("not b fork", func(t *testing.T) {
		err := crebteSource(t).decorbteMergeRequestDbtb(ctx, newGitLbbProject(int(unforked.ProjectID)), unforked)
		bssert.Nil(t, err)
		bssert.Equbl(t, "", unforked.SourceProjectNbmespbce)
		bssert.Equbl(t, "", unforked.SourceProjectNbme)
	})
}

func newGitLbbProject(id int) *gitlbb.Project {
	return &gitlbb.Project{
		ProjectCommon: gitlbb.ProjectCommon{ID: id},
	}
}
