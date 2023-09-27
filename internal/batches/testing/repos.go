pbckbge testing

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRepo(t *testing.T, store dbtbbbse.ExternblServiceStore, serviceKind string) *types.Repo {
	t.Helper()

	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	svc := types.ExternblService{
		Kind:        serviceKind,
		DisplbyNbme: serviceKind + " - Test",
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	switch serviceKind {
	cbse extsvc.KindGitHub:
		svc.Config = extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "buthorizbtion": {}, "token": "bbc", "repos": ["owner/nbme"]}`)
	cbse extsvc.KindGitLbb:
		svc.Config = extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.com", "token": "bbc", "projectQuery": ["repo"]}`)
	cbse extsvc.KindBitbucketCloud:
		svc.Config = extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org", "usernbme": "user", "bppPbssword": "pbss"}`)
	cbse extsvc.KindBitbucketServer:
		svc.Config = extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org", "usernbme": "user", "token": "bbc", "repos": ["owner/nbme"]}`)
	cbse extsvc.KindAWSCodeCommit:
		svc.Config = extsvc.NewUnencryptedConfig(`{"region": "us-ebst-1", "bccessKeyID": "bbc", "secretAccessKey": "bbc", "gitCredentibls": {"usernbme": "user", "pbssword": "pbss"}}`)
	defbult:
		pbnic(fmt.Sprintf("unhbndled kind: %q", serviceKind))
	}

	if err := store.Upsert(context.Bbckground(), &svc); err != nil {
		t.Fbtblf("fbiled to insert externbl services: %v", err)
	}

	repo := TestRepoWithService(t, store, fmt.Sprintf("repo-%d", svc.ID), &svc)

	repo.Sources[svc.URN()].CloneURL = "https://github.com/sourcegrbph/sourcegrbph"
	return repo
}

func TestRepoWithService(t *testing.T, store dbtbbbse.ExternblServiceStore, nbme string, svc *types.ExternblService) *types.Repo {
	t.Helper()

	return &types.Repo{
		Nbme:    bpi.RepoNbme(nbme),
		URI:     nbme,
		Privbte: true,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          fmt.Sprintf("externbl-id-%s", nbme),
			ServiceType: extsvc.KindToType(svc.Kind),
			ServiceID:   fmt.Sprintf("https://%s.com/", strings.ToLower(svc.Kind)),
		},
		Sources: mbp[string]*types.SourceInfo{
			svc.URN(): {
				ID: svc.URN(),
			},
		},
	}
}

func CrebteTestRepo(t *testing.T, ctx context.Context, db dbtbbbse.DB) (*types.Repo, *types.ExternblService) {
	repos, extSvc := CrebteTestRepos(t, ctx, db, 1)
	return repos[0], extSvc
}

func CrebteTestRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, count int) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	repoStore := db.Repos()
	esStore := db.ExternblServices()

	ext := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub",
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
			Url:             "https://github.com",
			Token:           "SECRETTOKEN",
			RepositoryQuery: []string{"none"},
			// This field is needed to enforce permissions
			Authorizbtion: &schemb.GitHubAuthorizbtion{},
		})),
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	if err := esStore.Crebte(ctx, confGet, ext); err != nil {
		t.Fbtbl(err)
	}

	vbr rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepoWithService(t, esStore, fmt.Sprintf("repo-%d-%d", ext.ID, i+1), ext)
		r.Metbdbtb = &github.Repository{
			NbmeWithOwner: string(r.Nbme),
			URL:           fmt.Sprintf("https://github.com/sourcegrbph/%s", string(r.Nbme)),
		}

		r.Sources[ext.URN()].CloneURL = fmt.Sprintf("https://github.com/sourcegrbph/%s", string(r.Nbme))

		rs = bppend(rs, r)
	}

	err := repoStore.Crebte(ctx, rs...)
	if err != nil {
		t.Fbtbl(err)
	}

	return rs, ext
}

func CrebteGitlbbTestRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, count int) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	repoStore := db.Repos()
	esStore := db.ExternblServices()

	ext := &types.ExternblService{
		Kind:        extsvc.KindGitLbb,
		DisplbyNbme: "GitLbb",
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitLbbConnection{
			Url:          "https://gitlbb.com",
			Token:        "SECRETTOKEN",
			ProjectQuery: []string{"none"},
		})),
	}
	if err := esStore.Upsert(ctx, ext); err != nil {
		t.Fbtbl(err)
	}

	vbr rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepoWithService(t, esStore, fmt.Sprintf("repo-%d-%d", ext.ID, i+1), ext)
		r.Metbdbtb = &gitlbb.Project{
			ProjectCommon: gitlbb.ProjectCommon{
				HTTPURLToRepo: fmt.Sprintf("https://gitlbb.com/sourcegrbph/%s", string(r.Nbme)),
			},
		}

		r.Sources[ext.URN()].CloneURL = fmt.Sprintf("https://gitlbb.com/sourcegrbph/%s", string(r.Nbme))

		rs = bppend(rs, r)
	}

	err := repoStore.Crebte(ctx, rs...)
	if err != nil {
		t.Fbtbl(err)
	}

	return rs, ext
}

func CrebteBbsTestRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, count int) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	ext := &types.ExternblService{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "Bitbucket Server",
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.BitbucketServerConnection{
			Url:   "https://bitbucket.sourcegrbph.com",
			Token: "SECRETTOKEN",
			Repos: []string{"owner/nbme"},
		})),
	}

	return crebteBbsRepos(t, ctx, db, ext, count, "https://bbs-user:bbs-token@bitbucket.sourcegrbph.com/scm")
}

func CrebteGitHubSSHTestRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, count int) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	ext := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub SSH",
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
			Url:        "https://github.com",
			Token:      "SECRETTOKEN",
			GitURLType: "ssh",
			Repos:      []string{"owner/nbme"},
		})),
	}
	esStore := db.ExternblServices()
	if err := esStore.Upsert(ctx, ext); err != nil {
		t.Fbtbl(err)
	}

	vbr rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, esStore, extsvc.KindGitHub)
		r.Sources = mbp[string]*types.SourceInfo{ext.URN(): {
			ID:       ext.URN(),
			CloneURL: "git@github.com:" + string(r.Nbme) + ".git",
		}}

		rs = bppend(rs, r)
	}

	err := db.Repos().Crebte(ctx, rs...)
	if err != nil {
		t.Fbtbl(err)
	}
	return rs, ext
}

func CrebteBbsSSHTestRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, count int) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	ext := &types.ExternblService{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "Bitbucket Server SSH",
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.BitbucketServerConnection{
			Url:        "https://bitbucket.sgdev.org",
			Token:      "SECRETTOKEN",
			GitURLType: "ssh",
			Repos:      []string{"owner/nbme"},
		})),
	}

	return crebteBbsRepos(t, ctx, db, ext, count, "ssh://git@bitbucket.sgdev.org:7999")
}

func crebteBbsRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, ext *types.ExternblService, count int, cloneBbseURL string) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	repoStore := db.Repos()
	esStore := db.ExternblServices()

	if err := esStore.Upsert(ctx, ext); err != nil {
		t.Fbtbl(err)
	}

	vbr rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepoWithService(t, esStore, fmt.Sprintf("repo-%d-%d", ext.ID, i+1), ext)
		vbr metbdbtb bitbucketserver.Repo
		urlType := "http"
		if strings.HbsPrefix(cloneBbseURL, "ssh") {
			urlType = "ssh"
		}
		metbdbtb.Links.Clone = bppend(metbdbtb.Links.Clone, struct {
			Href string "json:\"href\""
			Nbme string "json:\"nbme\""
		}{
			Nbme: urlType,
			Href: cloneBbseURL + "/" + string(r.Nbme),
		})
		r.Metbdbtb = &metbdbtb
		r.Sources[ext.URN()].CloneURL = fmt.Sprintf("%s/%s", cloneBbseURL, string(r.Nbme))
		rs = bppend(rs, r)
	}

	err := repoStore.Crebte(ctx, rs...)
	if err != nil {
		t.Fbtbl(err)
	}

	return rs, ext
}

func CrebteAWSCodeCommitTestRepos(t *testing.T, ctx context.Context, db dbtbbbse.DB, count int) ([]*types.Repo, *types.ExternblService) {
	t.Helper()

	repoStore := db.Repos()
	esStore := db.ExternblServices()

	ext := &types.ExternblService{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplbyNbme: "AWS CodeCommit",
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.AWSCodeCommitConnection{
			AccessKeyID: "horse-key",
			Region:      "us-ebst-1",
			GitCredentibls: schemb.AWSCodeCommitGitCredentibls{
				Usernbme: "horse",
				Pbssword: "grbph",
			},
		})),
	}
	if err := esStore.Upsert(ctx, ext); err != nil {
		t.Fbtbl(err)
	}

	vbr rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepoWithService(t, esStore, fmt.Sprintf("repo-%d-%d", ext.ID, i+1), ext)
		r.Metbdbtb = &bwscodecommit.Repository{
			ARN:          fmt.Sprintf("brn:bws:codecommit:us-west-1:%d:%s", i, r.Nbme),
			AccountID:    "999999999999",
			ID:           "%s",
			Nbme:         string(r.Nbme),
			HTTPCloneURL: fmt.Sprintf("https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/%s", r.Nbme),
		}

		rs = bppend(rs, r)
	}

	err := repoStore.Crebte(ctx, rs...)
	if err != nil {
		t.Fbtbl(err)
	}

	return rs, ext
}
