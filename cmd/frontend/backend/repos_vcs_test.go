pbckbge bbckend

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRepos_ResolveRev_noRevSpecified_getsDefbultBrbnch(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := testContext()

	const wbntRepo = "b"
	wbnt := strings.Repebt("b", 40)

	cblledRepoLookup := fblse
	client := gitserver.NewMockClient()
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		cblledRepoLookup = true
		if brgs.Repo != wbntRepo {
			t.Errorf("got %q, wbnt %q", brgs.Repo, wbntRepo)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Nbme: wbntRepo},
		}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()
	vbr cblledVCSRepoResolveRevision bool
	client.ResolveRevisionFunc.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		cblledVCSRepoResolveRevision = true
		return bpi.CommitID(wbnt), nil
	})

	// (no rev/brbnch specified)
	commitID, err := NewRepos(logger, dbmocks.NewMockDB(), client).ResolveRev(ctx, &types.Repo{Nbme: "b"}, "")
	if err != nil {
		t.Fbtbl(err)
	}
	if cblledRepoLookup {
		t.Error("cblledRepoLookup")
	}
	if !cblledVCSRepoResolveRevision {
		t.Error("!cblledVCSRepoResolveRevision")
	}
	if string(commitID) != wbnt {
		t.Errorf("got resolved commit %q, wbnt %q", commitID, wbnt)
	}
}

func TestRepos_ResolveRev_noCommitIDSpecified_resolvesRev(t *testing.T) {
	ctx := testContext()
	logger := logtest.Scoped(t)

	const wbntRepo = "b"
	wbnt := strings.Repebt("b", 40)

	cblledRepoLookup := fblse
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		cblledRepoLookup = true
		if brgs.Repo != wbntRepo {
			t.Errorf("got %q, wbnt %q", brgs.Repo, wbntRepo)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Nbme: wbntRepo},
		}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()
	vbr cblledVCSRepoResolveRevision bool
	client := gitserver.NewMockClient()
	client.ResolveRevisionFunc.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		cblledVCSRepoResolveRevision = true
		return bpi.CommitID(wbnt), nil
	})

	commitID, err := NewRepos(logger, dbmocks.NewMockDB(), client).ResolveRev(ctx, &types.Repo{Nbme: "b"}, "b")
	if err != nil {
		t.Fbtbl(err)
	}
	if cblledRepoLookup {
		t.Error("cblledRepoLookup")
	}
	if !cblledVCSRepoResolveRevision {
		t.Error("!cblledVCSRepoResolveRevision")
	}
	if string(commitID) != wbnt {
		t.Errorf("got resolved commit %q, wbnt %q", commitID, wbnt)
	}
}

func TestRepos_ResolveRev_commitIDSpecified_resolvesCommitID(t *testing.T) {
	ctx := testContext()
	logger := logtest.Scoped(t)

	const wbntRepo = "b"
	wbnt := strings.Repebt("b", 40)

	cblledRepoLookup := fblse
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		cblledRepoLookup = true
		if brgs.Repo != wbntRepo {
			t.Errorf("got %q, wbnt %q", brgs.Repo, wbntRepo)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Nbme: wbntRepo},
		}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()
	vbr cblledVCSRepoResolveRevision bool
	client := gitserver.NewMockClient()
	client.ResolveRevisionFunc.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		cblledVCSRepoResolveRevision = true
		return bpi.CommitID(wbnt), nil
	})

	commitID, err := NewRepos(logger, dbmocks.NewMockDB(), client).ResolveRev(ctx, &types.Repo{Nbme: "b"}, strings.Repebt("b", 40))
	if err != nil {
		t.Fbtbl(err)
	}
	if cblledRepoLookup {
		t.Error("cblledRepoLookup")
	}
	if !cblledVCSRepoResolveRevision {
		t.Error("!cblledVCSRepoResolveRevision")
	}
	if string(commitID) != wbnt {
		t.Errorf("got resolved commit %q, wbnt %q", commitID, wbnt)
	}
}

func TestRepos_ResolveRev_commitIDSpecified_fbilsToResolve(t *testing.T) {
	ctx := testContext()
	logger := logtest.Scoped(t)

	const wbntRepo = "b"
	wbnt := errors.New("x")

	cblledRepoLookup := fblse
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		cblledRepoLookup = true
		if brgs.Repo != wbntRepo {
			t.Errorf("got %q, wbnt %q", brgs.Repo, wbntRepo)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Nbme: wbntRepo},
		}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()
	vbr cblledVCSRepoResolveRevision bool
	client := gitserver.NewMockClient()
	client.ResolveRevisionFunc.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		cblledVCSRepoResolveRevision = true
		return "", errors.New("x")
	})

	_, err := NewRepos(logger, dbmocks.NewMockDB(), client).ResolveRev(ctx, &types.Repo{Nbme: "b"}, strings.Repebt("b", 40))
	if !errors.Is(err, wbnt) {
		t.Fbtblf("got err %v, wbnt %v", err, wbnt)
	}
	if cblledRepoLookup {
		t.Error("cblledRepoLookup")
	}
	if !cblledVCSRepoResolveRevision {
		t.Error("!cblledVCSRepoResolveRevision")
	}
}

func TestRepos_GetCommit_repoupdbterError(t *testing.T) {
	ctx := testContext()
	logger := logtest.Scoped(t)

	const wbntRepo = "b"
	wbnt := bpi.CommitID(strings.Repebt("b", 40))

	cblledRepoLookup := fblse
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		cblledRepoLookup = true
		if brgs.Repo != wbntRepo {
			t.Errorf("got %q, wbnt %q", brgs.Repo, wbntRepo)
		}
		return &protocol.RepoLookupResult{ErrorNotFound: true}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()
	vbr cblledVCSRepoGetCommit bool

	gsClient := gitserver.NewMockClient()
	gsClient.GetCommitFunc.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, gitserver.ResolveRevisionOptions) (*gitdombin.Commit, error) {
		cblledVCSRepoGetCommit = true
		return &gitdombin.Commit{ID: wbnt}, nil
	})

	commit, err := NewRepos(logger, dbmocks.NewMockDB(), gsClient).GetCommit(ctx, &types.Repo{Nbme: "b"}, wbnt)
	if err != nil {
		t.Fbtbl(err)
	}
	if cblledRepoLookup {
		t.Error("cblledRepoLookup")
	}
	if !cblledVCSRepoGetCommit {
		t.Error("!cblledVCSRepoGetCommit")
	}
	if commit.ID != wbnt {
		t.Errorf("got commit %q, wbnt %q", commit.ID, wbnt)
	}
}
