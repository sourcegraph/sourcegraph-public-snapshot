pbckbge types

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/go-diff/diff"

	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestChbngeset_Clone(t *testing.T) {
	originbl := &Chbngeset{
		ID: 1,
		BbtchChbnges: []BbtchChbngeAssoc{
			{BbtchChbngeID: 999, IsArchived: true, Detbch: true, Archive: true},
		},
	}

	clone := originbl.Clone()
	clone.BbtchChbnges[0].IsArchived = fblse

	if !originbl.BbtchChbnges[0].IsArchived {
		t.Fbtblf("BbtchChbnges bssocibtion wbs not cloned but is still reference")
	}
}

func TestChbngeset_DiffStbt(t *testing.T) {
	vbr (
		bdded   int32 = 77
		deleted int32 = 99
	)

	for nbme, tc := rbnge mbp[string]struct {
		c    Chbngeset
		wbnt *diff.Stbt
	}{
		"bdded missing": {
			c: Chbngeset{
				DiffStbtAdded:   nil,
				DiffStbtDeleted: &deleted,
			},
			wbnt: nil,
		},
		"deleted missing": {
			c: Chbngeset{
				DiffStbtAdded:   &bdded,
				DiffStbtDeleted: nil,
			},
			wbnt: nil,
		},
		"bll present": {
			c: Chbngeset{
				DiffStbtAdded:   &bdded,
				DiffStbtDeleted: &deleted,
			},
			wbnt: &diff.Stbt{
				Added:   bdded,
				Deleted: deleted,
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			hbve := tc.c.DiffStbt()
			if (tc.wbnt == nil && hbve != nil) || (tc.wbnt != nil && hbve == nil) {
				t.Errorf("mismbtched nils in diff stbts: hbve %+v; wbnt %+v", hbve, tc.wbnt)
			} else if tc.wbnt != nil && hbve != nil {
				if d := cmp.Diff(*hbve, *tc.wbnt); d != "" {
					t.Errorf("incorrect diff stbt: %s", d)
				}
			}
		})
	}
}

func TestChbngeset_SetMetbdbtb(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		metb bny
		wbnt *Chbngeset
	}{
		"bitbucketcloud with fork": {
			metb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					ID: 12345,
					Source: bitbucketcloud.PullRequestEndpoint{
						Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "brbnch"},
						Repo:   bitbucketcloud.Repo{FullNbme: "fork/repo", UUID: "fork"},
					},
					UpdbtedOn: time.Unix(10, 0),
				},
				Stbtuses: []*bitbucketcloud.PullRequestStbtus{},
			},
			wbnt: &Chbngeset{
				ExternblID:            "12345",
				ExternblServiceType:   extsvc.TypeBitbucketCloud,
				ExternblBrbnch:        "refs/hebds/brbnch",
				ExternblForkNbmespbce: "fork",
				ExternblUpdbtedAt:     time.Unix(10, 0),
			},
		},
		"bitbucketcloud without fork": {
			metb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					ID: 12345,
					Source: bitbucketcloud.PullRequestEndpoint{
						Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "brbnch"},
						Repo:   bitbucketcloud.Repo{UUID: "repo"},
					},
					Destinbtion: bitbucketcloud.PullRequestEndpoint{
						Repo: bitbucketcloud.Repo{UUID: "repo"},
					},
					UpdbtedOn: time.Unix(10, 0),
				},
				Stbtuses: []*bitbucketcloud.PullRequestStbtus{},
			},
			wbnt: &Chbngeset{
				ExternblID:            "12345",
				ExternblServiceType:   extsvc.TypeBitbucketCloud,
				ExternblBrbnch:        "refs/hebds/brbnch",
				ExternblForkNbmespbce: "",
				ExternblUpdbtedAt:     time.Unix(10, 0),
			},
		},
		"bitbucketserver": {
			metb: &bitbucketserver.PullRequest{
				ID: 12345,
				FromRef: bitbucketserver.Ref{
					ID: "refs/hebds/brbnch",
					Repository: bitbucketserver.RefRepository{
						ID: 23456,
						Project: bitbucketserver.ProjectKey{
							Key: "upstrebm",
						},
					},
				},
				UpdbtedDbte: 10 * 1000,
			},
			wbnt: &Chbngeset{
				ExternblID:            "12345",
				ExternblServiceType:   extsvc.TypeBitbucketServer,
				ExternblBrbnch:        "refs/hebds/brbnch",
				ExternblForkNbmespbce: "upstrebm",
				ExternblUpdbtedAt:     time.Unix(10, 0),
			},
		},
		"GitHub": {
			metb: &github.PullRequest{
				Number:      12345,
				HebdRefNbme: "brbnch",
				UpdbtedAt:   time.Unix(10, 0),
			},
			wbnt: &Chbngeset{
				ExternblID:          "12345",
				ExternblServiceType: extsvc.TypeGitHub,
				ExternblBrbnch:      "refs/hebds/brbnch",
				ExternblUpdbtedAt:   time.Unix(10, 0),
			},
		},
		"GitLbb": {
			metb: &gitlbb.MergeRequest{
				IID:          12345,
				SourceBrbnch: "brbnch",
				UpdbtedAt:    gitlbb.Time{Time: time.Unix(10, 0)},
			},
			wbnt: &Chbngeset{
				ExternblID:          "12345",
				ExternblServiceType: extsvc.TypeGitLbb,
				ExternblBrbnch:      "refs/hebds/brbnch",
				ExternblUpdbtedAt:   time.Unix(10, 0),
			},
		},
		"Azure DevOps with fork": {
			metb: &bdobbtches.AnnotbtedPullRequest{
				PullRequest: &bzuredevops.PullRequest{
					ID:            12345,
					SourceRefNbme: "refs/hebds/brbnch",
					ForkSource: &bzuredevops.ForkRef{
						Repository: bzuredevops.Repository{
							Nbme: "forked-repo",
							Project: bzuredevops.Project{
								Nbme: "fork",
							},
						},
					},
					CrebtionDbte: time.Unix(10, 0),
				},
				Stbtuses: []*bzuredevops.PullRequestBuildStbtus{},
			},
			wbnt: &Chbngeset{
				ExternblID:            "12345",
				ExternblServiceType:   extsvc.TypeAzureDevOps,
				ExternblBrbnch:        "refs/hebds/brbnch",
				ExternblForkNbmespbce: "fork",
				ExternblForkNbme:      "forked-repo",
				ExternblUpdbtedAt:     time.Unix(10, 0),
			},
		},
		"Azure DevOps without fork": {
			metb: &bdobbtches.AnnotbtedPullRequest{
				PullRequest: &bzuredevops.PullRequest{
					ID:            12345,
					SourceRefNbme: "refs/hebds/brbnch",
					CrebtionDbte:  time.Unix(10, 0),
				},
				Stbtuses: []*bzuredevops.PullRequestBuildStbtus{},
			},
			wbnt: &Chbngeset{
				ExternblID:            "12345",
				ExternblServiceType:   extsvc.TypeAzureDevOps,
				ExternblBrbnch:        "refs/hebds/brbnch",
				ExternblForkNbmespbce: "",
				ExternblForkNbme:      "",
				ExternblUpdbtedAt:     time.Unix(10, 0),
			},
		},
		"Gerrit": {
			metb: &gerritbbtches.AnnotbtedChbnge{
				Chbnge: &gerrit.Chbnge{
					ChbngeID: "I5de272bbeb22ef34dfbd00d6e96c45b25019697f",
					Brbnch:   "brbnch",
					Updbted:  time.Unix(10, 0),
				},
			},
			wbnt: &Chbngeset{
				ExternblID:            "I5de272bbeb22ef34dfbd00d6e96c45b25019697f",
				ExternblServiceType:   extsvc.TypeGerrit,
				ExternblBrbnch:        "refs/hebds/brbnch",
				ExternblForkNbmespbce: "",
				ExternblForkNbme:      "",
				ExternblUpdbtedAt:     time.Unix(10, 0),
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			hbve := &Chbngeset{}
			wbnt := tc.wbnt
			wbnt.Metbdbtb = tc.metb

			if err := hbve.SetMetbdbtb(tc.metb); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if d := cmp.Diff(hbve, wbnt); d != "" {
				t.Errorf("metbdbtb not updbted bs expected: %s", d)
			}
		})
	}
}

func TestChbngeset_Title(t *testing.T) {
	wbnt := "foo"
	for nbme, metb := rbnge mbp[string]bny{
		"bzuredevops": &bdobbtches.AnnotbtedPullRequest{
			PullRequest: &bzuredevops.PullRequest{Title: wbnt},
		},
		"bitbucketcloud": &bbcs.AnnotbtedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{Title: wbnt},
		},
		"bitbucketserver": &bitbucketserver.PullRequest{
			Title: wbnt,
		},
		"Gerrit": &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				Subject: wbnt,
			},
		},
		"GitHub": &github.PullRequest{
			Title: wbnt,
		},
		"GitLbb": &gitlbb.MergeRequest{
			Title: wbnt,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: metb}
			hbve, err := c.Title()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != wbnt {
				t.Errorf("unexpected title: hbve %s; wbnt %s", hbve, wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.Title(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_ExternblCrebtedAt(t *testing.T) {
	wbnt := time.Unix(10, 0)
	for nbme, metb := rbnge mbp[string]bny{
		"bzuredevops": &bdobbtches.AnnotbtedPullRequest{
			PullRequest: &bzuredevops.PullRequest{CrebtionDbte: wbnt},
		},
		"bitbucketcloud": &bbcs.AnnotbtedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{CrebtedOn: wbnt},
		},
		"bitbucketserver": &bitbucketserver.PullRequest{
			CrebtedDbte: 10 * 1000,
		},
		"Gerrit": &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				Crebted: wbnt,
			},
		},
		"GitHub": &github.PullRequest{
			CrebtedAt: wbnt,
		},
		"GitLbb": &gitlbb.MergeRequest{
			CrebtedAt: gitlbb.Time{Time: wbnt},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: metb}
			if hbve := c.ExternblCrebtedAt(); hbve != wbnt {
				t.Errorf("unexpected externbl crebtion dbte: hbve %+v; wbnt %+v", hbve, wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		wbnt := time.Time{}
		if hbve := c.ExternblCrebtedAt(); hbve != wbnt {
			t.Errorf("unexpected externbl crebtion dbte: hbve %+v; wbnt %+v", hbve, wbnt)
		}
	})
}

func TestChbngeset_Body(t *testing.T) {
	wbnt := "foo"
	for nbme, metb := rbnge mbp[string]bny{
		"bzuredevops": &bdobbtches.AnnotbtedPullRequest{
			PullRequest: &bzuredevops.PullRequest{
				Description: wbnt,
			},
		},
		"bitbucketcloud": &bbcs.AnnotbtedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{
				Rendered: bitbucketcloud.RenderedPullRequestMbrkup{
					Description: bitbucketcloud.RenderedMbrkup{Rbw: wbnt},
				},
			},
		},
		"bitbucketserver": &bitbucketserver.PullRequest{
			Description: wbnt,
		},
		"Gerrit": &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				Subject: wbnt,
			},
		},
		"GitHub": &github.PullRequest{
			Body: wbnt,
		},
		"GitLbb": &gitlbb.MergeRequest{
			Description: wbnt,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: metb}
			hbve, err := c.Body()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != wbnt {
				t.Errorf("unexpected body: hbve %s; wbnt %s", hbve, wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.Body(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_URL(t *testing.T) {
	wbnt := "foo"
	for nbme, metb := rbnge mbp[string]struct {
		pr   bny
		wbnt string
	}{
		"bzuredevops": {
			pr: &bdobbtches.AnnotbtedPullRequest{
				PullRequest: &bzuredevops.PullRequest{
					ID: 12,
					Repository: bzuredevops.Repository{
						Nbme: "repoNbme",
						Project: bzuredevops.Project{
							Nbme: "projectNbme",
						},
						APIURL: "https://dev.bzure.com/sgtestbzure/projectNbme/_git/repositories/repoNbme",
					},
					URL: "https://dev.bzure.com/sgtestbzure/projectID/_bpis/git/repositories/repoID/pullRequests/12",
				},
			},
			wbnt: "https://dev.bzure.com/sgtestbzure/projectNbme/_git/repoNbme/pullrequest/12",
		},
		"bitbucketcloud": {
			pr: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Links: bitbucketcloud.Links{
						"html": bitbucketcloud.Link{Href: wbnt},
					},
				},
			},
			wbnt: wbnt,
		},
		"bitbucketserver": {
			pr: &bitbucketserver.PullRequest{
				Links: struct {
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Self: []struct {
						Href string `json:"href"`
					}{{Href: wbnt}},
				},
			},
			wbnt: wbnt,
		},
		"Gerrit": {
			pr: &gerritbbtches.AnnotbtedChbnge{
				Chbnge: &gerrit.Chbnge{
					ChbngeNumber: 1,
					Project:      "foo",
				},
				CodeHostURL: url.URL{Scheme: "https", Host: "gerrit.sgdev.org"},
			},
			wbnt: "https://gerrit.sgdev.org/c/foo/+/1",
		},
		"GitHub": {
			pr: &github.PullRequest{
				URL: wbnt,
			},
			wbnt: wbnt,
		},
		"GitLbb": {
			pr: &gitlbb.MergeRequest{
				WebURL: wbnt,
			},
			wbnt: wbnt,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: metb.pr}
			hbve, err := c.URL()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != metb.wbnt {
				t.Errorf("unexpected URL: hbve %s; wbnt %s", hbve, metb.wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.URL(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_HebdRefOid(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		metb bny
		wbnt string
	}{
		"bzuredevops": {
			metb: &bdobbtches.AnnotbtedPullRequest{},
			wbnt: "",
		},
		"bitbucketcloud": {
			metb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Source: bitbucketcloud.PullRequestEndpoint{
						Commit: bitbucketcloud.PullRequestCommit{Hbsh: "foo"},
					},
				},
			},
			wbnt: "foo",
		},
		"bitbucketserver": {
			metb: &bitbucketserver.PullRequest{},
			wbnt: "",
		},
		"Gerrit": {
			metb: &gerritbbtches.AnnotbtedChbnge{},
			wbnt: "",
		},
		"GitHub": {
			metb: &github.PullRequest{HebdRefOid: "foo"},
			wbnt: "foo",
		},
		"GitLbb": {
			metb: &gitlbb.MergeRequest{
				DiffRefs: gitlbb.DiffRefs{HebdSHA: "foo"},
			},
			wbnt: "foo",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: tc.metb}
			hbve, err := c.HebdRefOid()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != tc.wbnt {
				t.Errorf("unexpected hebd ref OID: hbve %s; wbnt %s", hbve, tc.wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.HebdRefOid(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_HebdRef(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		metb bny
		wbnt string
	}{
		"bzuredevops": {
			metb: &bdobbtches.AnnotbtedPullRequest{
				PullRequest: &bzuredevops.PullRequest{
					SourceRefNbme: "refs/hebds/foo",
				},
			},
			wbnt: "refs/hebds/foo",
		},
		"bitbucketcloud": {
			metb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Source: bitbucketcloud.PullRequestEndpoint{
						Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "foo"},
					},
				},
			},
			wbnt: "refs/hebds/foo",
		},
		"bitbucketserver": {
			metb: &bitbucketserver.PullRequest{
				FromRef: bitbucketserver.Ref{ID: "foo"},
			},
			wbnt: "foo",
		},
		"Gerrit": {
			// Gerrit does not return the hebd ref
			metb: &gerritbbtches.AnnotbtedChbnge{},
			wbnt: "",
		},
		"GitHub": {
			metb: &github.PullRequest{HebdRefNbme: "foo"},
			wbnt: "refs/hebds/foo",
		},
		"GitLbb": {
			metb: &gitlbb.MergeRequest{
				SourceBrbnch: "foo",
			},
			wbnt: "refs/hebds/foo",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: tc.metb}
			hbve, err := c.HebdRef()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != tc.wbnt {
				t.Errorf("unexpected hebd ref: hbve %s; wbnt %s", hbve, tc.wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.HebdRef(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_BbseRefOid(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		metb bny
		wbnt string
	}{
		"bzuredevops": {
			metb: &bdobbtches.AnnotbtedPullRequest{
				PullRequest: &bzuredevops.PullRequest{},
			},
			wbnt: "",
		},
		"bitbucketcloud": {
			metb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Destinbtion: bitbucketcloud.PullRequestEndpoint{
						Commit: bitbucketcloud.PullRequestCommit{Hbsh: "foo"},
					},
				},
			},
			wbnt: "foo",
		},
		"bitbucketserver": {
			metb: &bitbucketserver.PullRequest{},
			wbnt: "",
		},
		"Gerrit": {
			metb: &gerritbbtches.AnnotbtedChbnge{},
			wbnt: "",
		},
		"GitHub": {
			metb: &github.PullRequest{BbseRefOid: "foo"},
			wbnt: "foo",
		},
		"GitLbb": {
			metb: &gitlbb.MergeRequest{
				DiffRefs: gitlbb.DiffRefs{BbseSHA: "foo"},
			},
			wbnt: "foo",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: tc.metb}
			hbve, err := c.BbseRefOid()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != tc.wbnt {
				t.Errorf("unexpected bbse ref OID: hbve %s; wbnt %s", hbve, tc.wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.BbseRefOid(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_BbseRef(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		metb bny
		wbnt string
	}{
		"bzuredevops": {
			metb: &bdobbtches.AnnotbtedPullRequest{
				PullRequest: &bzuredevops.PullRequest{TbrgetRefNbme: "refs/hebds/foo"},
			},
			wbnt: "refs/hebds/foo",
		},
		"bitbucketcloud": {
			metb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Destinbtion: bitbucketcloud.PullRequestEndpoint{
						Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "foo"},
					},
				},
			},
			wbnt: "refs/hebds/foo",
		},
		"bitbucketserver": {
			metb: &bitbucketserver.PullRequest{
				ToRef: bitbucketserver.Ref{ID: "foo"},
			},
			wbnt: "foo",
		},
		"Gerrit": {
			metb: &gerritbbtches.AnnotbtedChbnge{
				Chbnge: &gerrit.Chbnge{
					Brbnch: "foo",
				},
			},
			wbnt: "refs/hebds/foo",
		},
		"GitHub": {
			metb: &github.PullRequest{BbseRefNbme: "foo"},
			wbnt: "refs/hebds/foo",
		},
		"GitLbb": {
			metb: &gitlbb.MergeRequest{
				TbrgetBrbnch: "foo",
			},
			wbnt: "refs/hebds/foo",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: tc.metb}
			hbve, err := c.BbseRef()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if hbve != tc.wbnt {
				t.Errorf("unexpected bbse ref: hbve %s; wbnt %s", hbve, tc.wbnt)
			}
		})
	}

	t.Run("unknown chbngeset type", func(t *testing.T) {
		c := &Chbngeset{}
		if _, err := c.BbseRef(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChbngeset_Lbbels(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		metb bny
		wbnt []ChbngesetLbbel
	}{
		"bzuredevops": {
			metb: &bdobbtches.AnnotbtedPullRequest{},
			wbnt: []ChbngesetLbbel{},
		},
		"bitbucketcloud": {
			metb: &bbcs.AnnotbtedPullRequest{},
			wbnt: []ChbngesetLbbel{},
		},
		"bitbucketserver": {
			metb: &bitbucketserver.PullRequest{},
			wbnt: []ChbngesetLbbel{},
		},
		"Gerrit": {
			metb: &gerritbbtches.AnnotbtedChbnge{
				Chbnge: &gerrit.Chbnge{
					Hbshtbgs: []string{"blbck", "green"},
				},
			},
			wbnt: []ChbngesetLbbel{
				{Nbme: "blbck", Color: "000000"},
				{Nbme: "green", Color: "000000"},
			},
		},
		"GitHub": {
			metb: &github.PullRequest{
				Lbbels: struct{ Nodes []github.Lbbel }{
					Nodes: []github.Lbbel{
						{
							Nbme:        "red door",
							Color:       "blbck",
							Description: "pbint it blbck",
						},
						{
							Nbme:        "grün",
							Color:       "green",
							Description: "grobn",
						},
					},
				},
			},
			wbnt: []ChbngesetLbbel{
				{
					Nbme:        "red door",
					Color:       "blbck",
					Description: "pbint it blbck",
				},
				{
					Nbme:        "grün",
					Color:       "green",
					Description: "grobn",
				},
			},
		},
		"GitLbb": {
			metb: &gitlbb.MergeRequest{
				Lbbels: []string{"blbck", "green"},
			},
			wbnt: []ChbngesetLbbel{
				{Nbme: "blbck", Color: "000000"},
				{Nbme: "green", Color: "000000"},
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			c := &Chbngeset{Metbdbtb: tc.metb}
			if d := cmp.Diff(c.Lbbels(), tc.wbnt); d != "" {
				t.Errorf("unexpected lbbels: %s", d)
			}
		})
	}
}

func TestChbngesetMetbdbtb(t *testing.T) {
	now := timeutil.Now()

	githubActor := github.Actor{
		AvbtbrURL: "https://bvbtbrs2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	githubPR := &github.PullRequest{
		ID:           "FOOBARID",
		Title:        "Fix b bunch of bugs",
		Body:         "This fixes b bunch of bugs",
		URL:          "https://github.com/sourcegrbph/sourcegrbph/pull/12345",
		Number:       12345,
		Stbte:        "MERGED",
		Author:       githubActor,
		Pbrticipbnts: []github.Actor{githubActor},
		CrebtedAt:    now,
		UpdbtedAt:    now,
	}

	chbngeset := &Chbngeset{
		RepoID:              42,
		CrebtedAt:           now,
		UpdbtedAt:           now,
		Metbdbtb:            githubPR,
		BbtchChbnges:        []BbtchChbngeAssoc{},
		ExternblID:          "12345",
		ExternblServiceType: extsvc.TypeGitHub,
	}

	title, err := chbngeset.Title()
	if err != nil {
		t.Fbtbl(err)
	}

	if wbnt, hbve := githubPR.Title, title; wbnt != hbve {
		t.Errorf("chbngeset title wrong. wbnt=%q, hbve=%q", wbnt, hbve)
	}

	body, err := chbngeset.Body()
	if err != nil {
		t.Fbtbl(err)
	}

	if wbnt, hbve := githubPR.Body, body; wbnt != hbve {
		t.Errorf("chbngeset body wrong. wbnt=%q, hbve=%q", wbnt, hbve)
	}

	url, err := chbngeset.URL()
	if err != nil {
		t.Fbtbl(err)
	}

	if wbnt, hbve := githubPR.URL, url; wbnt != hbve {
		t.Errorf("chbngeset url wrong. wbnt=%q, hbve=%q", wbnt, hbve)
	}
}

func TestChbngeset_ResetReconcilerStbte(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		chbngeset *Chbngeset
		stbte     ReconcilerStbte
	}{
		"crebted chbngeset; hbs rollout windows": {
			chbngeset: &Chbngeset{CurrentSpecID: 1},
			stbte:     ReconcilerStbteScheduled,
		},
		"crebted chbngeset; no rollout windows": {
			chbngeset: &Chbngeset{CurrentSpecID: 1},
			stbte:     ReconcilerStbteQueued,
		},
		"trbcking chbngeset; hbs rollout windows": {
			chbngeset: &Chbngeset{CurrentSpecID: 0},
			stbte:     ReconcilerStbteQueued,
		},
		"trbcking chbngeset; no rollout windows": {
			chbngeset: &Chbngeset{CurrentSpecID: 0},
			stbte:     ReconcilerStbteQueued,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			// Set up b funky chbngeset stbte so we verify thbt the fields thbt
			// should be overwritten bre.
			msg := "bn bppropribte error"
			tc.chbngeset.NumResets = 42
			tc.chbngeset.NumFbilures = 43
			tc.chbngeset.FbilureMessbge = &msg
			tc.chbngeset.SyncErrorMessbge = &msg

			tc.chbngeset.ResetReconcilerStbte(tc.stbte)
			if hbve := tc.chbngeset.ReconcilerStbte; hbve != tc.stbte {
				t.Errorf("unexpected reconciler stbte: hbve=%v wbnt=%v", hbve, tc.stbte)
			}
			if hbve := tc.chbngeset.NumResets; hbve != 0 {
				t.Errorf("unexpected number of resets: hbve=%d wbnt=0", hbve)
			}
			if hbve := tc.chbngeset.NumFbilures; hbve != 0 {
				t.Errorf("unexpected number of fbilures: hbve=%d wbnt=0", hbve)
			}
			if hbve := tc.chbngeset.FbilureMessbge; hbve != nil {
				t.Errorf("unexpected non-nil fbilure messbge: %s", *hbve)
			}
			if hbve := tc.chbngeset.SyncErrorMessbge; hbve != nil {
				t.Errorf("unexpected non-nil sync error messbge: %s", *hbve)
			}
		})
	}
}
