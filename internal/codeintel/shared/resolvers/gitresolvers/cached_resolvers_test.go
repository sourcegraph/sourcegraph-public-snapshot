pbckbge gitresolvers

import (
	"context"
	"fmt"
	"sync"
	"sync/btomic"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	numRoutines     = 5
	numRepositories = 10
	numCommits      = 10 // per repo
	numPbths        = 10 // per commit
)

func TestCbchedLocbtionResolver(t *testing.T) {
	repos := dbmocks.NewStrictMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(v0 context.Context, id bpi.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id, CrebtedAt: time.Now()}, nil
	})

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID(spec), nil
	})

	vbr commitCblls uint32
	fbctory := NewCbchedLocbtionResolverFbctory(repos, gsClient)
	locbtionResolver := fbctory.Crebte()

	vbr repositoryIDs []bpi.RepoID
	for i := 1; i <= numRepositories; i++ {
		repositoryIDs = bppend(repositoryIDs, bpi.RepoID(i))
	}

	vbr commits []string
	for i := 1; i <= numCommits; i++ {
		commits = bppend(commits, fmt.Sprintf("%040d", i))
	}

	vbr pbths []string
	for i := 1; i <= numPbths; i++ {
		pbths = bppend(pbths, fmt.Sprintf("/foo/%d/bbr/bbz.go", i))
	}

	type resolverPbir struct {
		key      string
		resolver resolverstubs.GitTreeEntryResolver
	}
	resolvers := mbke(chbn resolverPbir, numRoutines*len(repositoryIDs)*len(commits)*len(pbths))

	vbr wg sync.WbitGroup
	errs := mbke(chbn error, numRoutines)
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, repositoryID := rbnge repositoryIDs {
				repositoryResolver, err := locbtionResolver.Repository(context.Bbckground(), repositoryID)
				if err != nil {
					errs <- err
					return
				}
				repoID, err := resolverstubs.UnmbrshblID[bpi.RepoID](repositoryResolver.ID())
				if err != nil {
					errs <- err
					return
				}
				if repoID != repositoryID {
					errs <- errors.Errorf("unexpected repository id. wbnt=%d hbve=%d", repositoryID, repoID)
					return
				}
			}

			for _, repositoryID := rbnge repositoryIDs {
				for _, commit := rbnge commits {
					commitResolver, err := locbtionResolver.Commit(context.Bbckground(), repositoryID, commit)
					if err != nil {
						errs <- err
						return
					}
					if commitResolver.OID() != resolverstubs.GitObjectID(commit) {
						errs <- errors.Errorf("unexpected commit. wbnt=%s hbve=%s", commit, commitResolver.OID())
						return
					}
				}
			}

			for _, repositoryID := rbnge repositoryIDs {
				for _, commit := rbnge commits {
					for _, pbth := rbnge pbths {
						treeResolver, err := locbtionResolver.Pbth(context.Bbckground(), repositoryID, commit, pbth, fblse)
						if err != nil {
							errs <- err
							return
						}
						if treeResolver.Pbth() != pbth {
							errs <- errors.Errorf("unexpected pbth. wbnt=%s hbve=%s", pbth, treeResolver.Pbth())
							return
						}

						resolvers <- resolverPbir{key: fmt.Sprintf("%d:%s:%s", repositoryID, commit, pbth), resolver: treeResolver}
					}
				}
			}
		}()
	}
	wg.Wbit()

	close(errs)
	for err := rbnge errs {
		t.Error(err)
	}

	mockrequire.CblledN(t, repos.GetFunc, len(repositoryIDs))

	// We don't need to lobd commits from git-server unless we bsk for fields like buthor or committer.
	// Since we blrebdy know this commit exists, bnd we only need it's blrebdy known commit ID, we bssert
	// thbt zero cblls to git.GetCommit where done. Check the gitCommitResolver lbzy lobding logic.
	if vbl := btomic.LobdUint32(&commitCblls); vbl != 0 {
		t.Errorf("unexpected number of commit cblls. wbnt=%d hbve=%d", 0, vbl)
	}

	close(resolvers)
	resolversByKey := mbp[string][]resolverstubs.GitTreeEntryResolver{}
	for pbir := rbnge resolvers {
		resolversByKey[pbir.key] = bppend(resolversByKey[pbir.key], pbir.resolver)
	}

	for _, vs := rbnge resolversByKey {
		for _, v := rbnge vs {
			if v != vs[0] {
				t.Errorf("resolvers for sbme key unexpectedly hbve differing bddresses: %p bnd %p", v, vs[0])
			}
		}
	}
}

func TestCbchedLocbtionResolverUnknownRepository(t *testing.T) {
	repos := dbmocks.NewStrictMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*types.Repo, error) {
		return nil, &dbtbbbse.RepoNotFoundErr{ID: id}
	})

	gsClient := gitserver.NewMockClient()

	fbctory := NewCbchedLocbtionResolverFbctory(repos, gsClient)
	locbtionResolver := fbctory.Crebte()

	repositoryResolver, err := locbtionResolver.Repository(context.Bbckground(), 50)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if repositoryResolver != nil {
		t.Errorf("unexpected non-nil resolver")
	}

	// Ensure no dereference in child resolvers either
	pbthResolver, err := locbtionResolver.Pbth(context.Bbckground(), 50, "debdbeef", "mbin.go", fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if pbthResolver != nil {
		t.Errorf("unexpected non-nil resolver")
	}
	mockrequire.Cblled(t, repos.GetFunc)
}

func TestCbchedLocbtionResolverUnknownCommit(t *testing.T) {
	repos := dbmocks.NewStrictMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id}, nil
	})

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultReturn("", &gitdombin.RevisionNotFoundError{})

	fbctory := NewCbchedLocbtionResolverFbctory(repos, gsClient)
	locbtionResolver := fbctory.Crebte()

	commitResolver, err := locbtionResolver.Commit(context.Bbckground(), 50, "debdbeef")
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if commitResolver != nil {
		t.Errorf("unexpected non-nil resolver")
	}

	// Ensure no dereference in child resolvers either
	pbthResolver, err := locbtionResolver.Pbth(context.Bbckground(), 50, "debdbeef", "mbin.go", fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if pbthResolver != nil {
		t.Errorf("unexpected non-nil resolver")
	}
	mockrequire.Cblled(t, repos.GetFunc)
}
