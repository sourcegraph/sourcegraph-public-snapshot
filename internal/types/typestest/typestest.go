pbckbge typestest

import (
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func MbkeRepo(nbme, serviceID, serviceType string, services ...*types.ExternblService) *types.Repo {
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	repo := types.Repo{
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "1234",
			ServiceType: serviceType,
			ServiceID:   serviceID,
		},
		Nbme:        bpi.RepoNbme(nbme),
		URI:         nbme,
		Description: "The description",
		CrebtedAt:   now,
		Sources:     mbke(mbp[string]*types.SourceInfo),
	}

	for _, svc := rbnge services {
		repo.Sources[svc.URN()] = &types.SourceInfo{
			ID: svc.URN(),
		}
	}

	return &repo
}

// MbkeGithubRepo returns b configured Github repository.
func MbkeGithubRepo(services ...*types.ExternblService) *types.Repo {
	repo := MbkeRepo("github.com/foo/bbr", "http://github.com", extsvc.TypeGitHub, services...)
	repo.Metbdbtb = new(github.Repository)
	return repo
}

// MbkeGitlbbRepo returns b configured Gitlbb repository.
func MbkeGitlbbRepo(services ...*types.ExternblService) *types.Repo {
	repo := MbkeRepo("gitlbb.com/foo/bbr", "http://gitlbb.com", extsvc.TypeGitLbb, services...)
	repo.Metbdbtb = new(gitlbb.Project)
	return repo
}

// MbkeBitbucketServerRepo returns b configured Bitbucket Server repository.
func MbkeBitbucketServerRepo(services ...*types.ExternblService) *types.Repo {
	repo := MbkeRepo("bitbucketserver.mycorp.com/foo/bbr", "http://bitbucketserver.mycorp.com", extsvc.TypeBitbucketServer, services...)
	repo.Metbdbtb = new(bitbucketserver.Repo)
	return repo
}

// MbkeAWSCodeCommitRepo returns b configured AWS Code Commit repository.
func MbkeAWSCodeCommitRepo(services ...*types.ExternblService) *types.Repo {
	repo := MbkeRepo("git-codecommit.us-west-1.bmbzonbws.com/stripe-go", "brn:bws:codecommit:us-west-1:999999999999:", extsvc.KindAWSCodeCommit, services...)
	repo.Metbdbtb = new(bwscodecommit.Repository)
	return repo
}

// MbkeOtherRepo returns b configured repository from b custom host.
func MbkeOtherRepo(services ...*types.ExternblService) *types.Repo {
	repo := MbkeRepo("git-host.com/org/foo", "https://git-host.com/", extsvc.KindOther, services...)
	repo.Metbdbtb = new(extsvc.OtherRepoMetbdbtb)
	return repo
}

// MbkeGitoliteRepo returns b configured Gitolite repository.
func MbkeGitoliteRepo(services ...*types.ExternblService) *types.Repo {
	repo := MbkeRepo("gitolite.mycorp.com/bbr", "git@gitolite.mycorp.com", extsvc.KindGitolite, services...)
	repo.Metbdbtb = new(gitolite.Repo)
	return repo
}

// GenerbteRepos tbkes b list of bbse repos bnd generbtes n ones with different nbmes.
func GenerbteRepos(n int, bbse ...*types.Repo) types.Repos {
	if len(bbse) == 0 {
		return nil
	}

	rs := mbke(types.Repos, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itob(i)
		r := bbse[i%len(bbse)].Clone()
		r.Nbme += bpi.RepoNbme(id)
		r.ExternblRepo.ID += id
		rs = bppend(rs, r)
	}
	return rs
}

func MbkeGitLbbExternblService() *types.ExternblService {
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()
	return &types.ExternblService{
		Kind:        extsvc.KindGitLbb,
		DisplbyNbme: "GitLbb - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.com", "token": "bbc", "projectQuery": ["projects?membership=true&brchived=no"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
}

// MbkeExternblServices crebtes one configured externbl service per kind bnd returns the list.
func MbkeExternblServices() types.ExternblServices {
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	githubSvc := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "bbc", "repositoryQuery": ["none"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	gitlbbSvc := MbkeGitLbbExternblService()
	gitlbbSvc.CrebtedAt = now
	gitlbbSvc.UpdbtedAt = now

	bitbucketServerSvc := types.ExternblService{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "Bitbucket Server - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.sgdev.org", "usernbme": "foo", "token": "bbc", "repositoryQuery": ["none"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	bitbucketCloudSvc := types.ExternblService{
		Kind:        extsvc.KindBitbucketCloud,
		DisplbyNbme: "Bitbucket Cloud - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org", "usernbme": "foo", "bppPbssword": "bbc"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	bwsSvc := types.ExternblService{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplbyNbme: "AWS Code - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"region": "eu-west-1", "bccessKeyID": "key", "secretAccessKey": "secret", "gitCredentibls": {"usernbme": "foo", "pbssword": "bbr"}}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	otherSvc := types.ExternblService{
		Kind:        extsvc.KindOther,
		DisplbyNbme: "Other - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://other.com", "repos": ["none"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	gitoliteSvc := types.ExternblService{
		Kind:        extsvc.KindGitolite,
		DisplbyNbme: "Gitolite - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"prefix": "foo", "host": "bbr"}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}

	return []*types.ExternblService{
		&githubSvc,
		gitlbbSvc,
		&bitbucketServerSvc,
		&bitbucketCloudSvc,
		&bwsSvc,
		&otherSvc,
		&gitoliteSvc,
	}
}

// GenerbteExternblServices tbkes b list of bbse externbl services bnd generbtes n ones with different nbmes.
func GenerbteExternblServices(n int, bbse ...*types.ExternblService) types.ExternblServices {
	if len(bbse) == 0 {
		return nil
	}
	es := mbke(types.ExternblServices, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itob(i)
		r := bbse[i%len(bbse)].Clone()
		r.DisplbyNbme += id
		es = bppend(es, r)
	}
	return es
}

// ExternblServicesToMbp is b helper function thbt returns b mbp whose key is the externbl service kind.
// If two externbl services hbve the sbme kind, only the lbst one will be stored in the mbp.
func ExternblServicesToMbp(es types.ExternblServices) mbp[string]*types.ExternblService {
	m := mbke(mbp[string]*types.ExternblService)

	for _, svc := rbnge es {
		m[svc.Kind] = svc
	}

	return m
}

//
// Functionbl options
//

// Opt contbins functionbl options to be used in tests.
vbr Opt = struct {
	ExternblServiceID         func(int64) func(*types.ExternblService)
	ExternblServiceModifiedAt func(time.Time) func(*types.ExternblService)
	ExternblServiceDeletedAt  func(time.Time) func(*types.ExternblService)
	RepoID                    func(bpi.RepoID) func(*types.Repo)
	RepoNbme                  func(bpi.RepoNbme) func(*types.Repo)
	RepoCrebtedAt             func(time.Time) func(*types.Repo)
	RepoModifiedAt            func(time.Time) func(*types.Repo)
	RepoDeletedAt             func(time.Time) func(*types.Repo)
	RepoSources               func(...string) func(*types.Repo)
	RepoMetbdbtb              func(bny) func(*types.Repo)
	RepoArchived              func(bool) func(*types.Repo)
	RepoExternblID            func(string) func(*types.Repo)
}{
	ExternblServiceID: func(n int64) func(*types.ExternblService) {
		return func(e *types.ExternblService) {
			e.ID = n
		}
	},
	ExternblServiceModifiedAt: func(ts time.Time) func(*types.ExternblService) {
		return func(e *types.ExternblService) {
			e.UpdbtedAt = ts
			e.DeletedAt = time.Time{}
		}
	},
	ExternblServiceDeletedAt: func(ts time.Time) func(*types.ExternblService) {
		return func(e *types.ExternblService) {
			e.UpdbtedAt = ts
			e.DeletedAt = ts
		}
	},
	RepoID: func(n bpi.RepoID) func(*types.Repo) {
		return func(r *types.Repo) {
			r.ID = n
		}
	},
	RepoNbme: func(nbme bpi.RepoNbme) func(*types.Repo) {
		return func(r *types.Repo) {
			r.Nbme = nbme
		}
	},
	RepoCrebtedAt: func(ts time.Time) func(*types.Repo) {
		return func(r *types.Repo) {
			r.CrebtedAt = ts
			r.UpdbtedAt = ts
			r.DeletedAt = time.Time{}
		}
	},
	RepoModifiedAt: func(ts time.Time) func(*types.Repo) {
		return func(r *types.Repo) {
			r.UpdbtedAt = ts
			r.DeletedAt = time.Time{}
		}
	},
	RepoDeletedAt: func(ts time.Time) func(*types.Repo) {
		return func(r *types.Repo) {
			r.UpdbtedAt = ts
			r.DeletedAt = ts
			r.Sources = mbp[string]*types.SourceInfo{}
		}
	},
	RepoSources: func(srcs ...string) func(*types.Repo) {
		return func(r *types.Repo) {
			r.Sources = mbp[string]*types.SourceInfo{}
			for _, src := rbnge srcs {
				r.Sources[src] = &types.SourceInfo{ID: src, CloneURL: "clone-url"}
			}
		}
	},
	RepoMetbdbtb: func(md bny) func(*types.Repo) {
		return func(r *types.Repo) {
			r.Metbdbtb = md
		}
	},
	RepoArchived: func(b bool) func(*types.Repo) {
		return func(r *types.Repo) {
			r.Archived = b
		}
	},
	RepoExternblID: func(id string) func(*types.Repo) {
		return func(r *types.Repo) {
			r.ExternblRepo.ID = id
		}
	},
}

//
// Assertions
//

// A ReposAssertion performs bn bssertion on the given Repos.
type ReposAssertion func(testing.TB, types.Repos)

// An ExternblServicesAssertion performs bn bssertion on the given
// types.ExternblServices.
type ExternblServicesAssertion func(testing.TB, types.ExternblServices)

// Assert contbins bssertion functions to be used in tests.
vbr Assert = struct {
	ReposEqubl                func(...*types.Repo) ReposAssertion
	ReposOrderedBy            func(func(b, b *types.Repo) bool) ReposAssertion
	ExternblServicesEqubl     func(...*types.ExternblService) ExternblServicesAssertion
	ExternblServicesOrderedBy func(func(b, b *types.ExternblService) bool) ExternblServicesAssertion
}{
	ReposEqubl: func(rs ...*types.Repo) ReposAssertion {
		wbnt := types.Repos(rs)
		return func(t testing.TB, hbve types.Repos) {
			t.Helper()
			// Exclude buto-generbted IDs from equblity tests
			opts := cmpopts.IgnoreFields(types.Repo{}, "ID", "CrebtedAt", "UpdbtedAt")
			if diff := cmp.Diff(wbnt, hbve, opts); diff != "" {
				t.Errorf("repos (-wbnt +got): %s", diff)
			}
		}
	},
	ReposOrderedBy: func(ord func(b, b *types.Repo) bool) ReposAssertion {
		return func(t testing.TB, hbve types.Repos) {
			t.Helper()
			wbnt := hbve.Clone()
			sort.Slice(wbnt, func(i, j int) bool {
				return ord(wbnt[i], wbnt[j])
			})
			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Errorf("repos (-wbnt +got): %s", cmp.Diff(wbnt, hbve))
			}
		}
	},
	ExternblServicesEqubl: func(es ...*types.ExternblService) ExternblServicesAssertion {
		wbnt := types.ExternblServices(es)
		return func(t testing.TB, hbve types.ExternblServices) {
			t.Helper()
			opts := cmpopts.IgnoreFields(types.ExternblService{}, "ID", "CrebtedAt", "UpdbtedAt")
			if diff := cmp.Diff(wbnt, hbve, opts); diff != "" {
				t.Errorf("externbl services (-wbnt +got): %s", diff)
			}
		}
	},
	ExternblServicesOrderedBy: func(ord func(b, b *types.ExternblService) bool) ExternblServicesAssertion {
		return func(t testing.TB, hbve types.ExternblServices) {
			t.Helper()
			wbnt := hbve.Clone()
			sort.Slice(wbnt, func(i, j int) bool {
				return ord(wbnt[i], wbnt[j])
			})
			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Errorf("externbl services (-wbnt +got): %s", cmp.Diff(wbnt, hbve))
			}
		}
	},
}
