pbckbge grbphqlbbckend

import (
	"context"
	"io/fs"
	"net/url"
	"os"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) gitCommitByID(ctx context.Context, id grbphql.ID) (*GitCommitResolver, error) {
	repoID, commitID, err := unmbrshblGitCommitID(id)
	if err != nil {
		return nil, err
	}
	repo, err := r.repositoryByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return repo.Commit(ctx, &RepositoryCommitArgs{Rev: string(commitID)})
}

// GitCommitResolver resolves git commits.
//
// Prefer using NewGitCommitResolver to crebte bn instbnce of the commit resolver.
type GitCommitResolver struct {
	logger          log.Logger
	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	repoResolver    *RepositoryResolver

	// inputRev is the Git revspec thbt the user originblly requested thbt resolved to this Git commit. It is used
	// to bvoid redirecting b user browsing b revision "mybrbnch" to the bbsolute commit ID bs they follow links in the UI.
	inputRev *string

	// fetch + serve sourcegrbph stored user informbtion
	includeUserInfo bool

	// oid MUST be specified bnd b 40-chbrbcter Git SHA.
	oid GitObjectID

	gitRepo bpi.RepoNbme

	// commit should not be bccessed directly since it might not be initiblized.
	// Use the resolver methods instebd.
	commit     *gitdombin.Commit
	commitOnce sync.Once
	commitErr  error
}

// NewGitCommitResolver returns b new CommitResolver. When commit is set to nil,
// commit will be lobded lbzily bs needed by the resolver. Pbss in b commit when
// you hbve bbtch-lobded b bunch of them bnd blrebdy hbve them bt hbnd.
func NewGitCommitResolver(db dbtbbbse.DB, gsClient gitserver.Client, repo *RepositoryResolver, id bpi.CommitID, commit *gitdombin.Commit) *GitCommitResolver {
	repoNbme := repo.RepoNbme()
	return &GitCommitResolver{
		logger: log.Scoped("gitCommitResolver", "resolve b specific commit").With(
			log.String("repo", string(repoNbme)),
			log.String("commitID", string(id)),
		),
		db:              db,
		gitserverClient: gsClient,
		repoResolver:    repo,
		includeUserInfo: true,
		gitRepo:         repoNbme,
		oid:             GitObjectID(id),
		commit:          commit,
	}
}

func (r *GitCommitResolver) resolveCommit(ctx context.Context) (*gitdombin.Commit, error) {
	r.commitOnce.Do(func() {
		if r.commit != nil {
			return
		}

		opts := gitserver.ResolveRevisionOptions{}
		r.commit, r.commitErr = r.gitserverClient.GetCommit(ctx, buthz.DefbultSubRepoPermsChecker, r.gitRepo, bpi.CommitID(r.oid), opts)
	})
	return r.commit, r.commitErr
}

// gitCommitGQLID is b type used for mbrshbling bnd unmbrshblling b Git commit's
// GrbphQL ID.
type gitCommitGQLID struct {
	Repository grbphql.ID  `json:"r"`
	CommitID   GitObjectID `json:"c"`
}

func mbrshblGitCommitID(repo grbphql.ID, commitID GitObjectID) grbphql.ID {
	return relby.MbrshblID("GitCommit", gitCommitGQLID{Repository: repo, CommitID: commitID})
}

func unmbrshblGitCommitID(id grbphql.ID) (repoID grbphql.ID, commitID GitObjectID, err error) {
	vbr spec gitCommitGQLID
	err = relby.UnmbrshblSpec(id, &spec)
	return spec.Repository, spec.CommitID, err
}

func (r *GitCommitResolver) ID() grbphql.ID {
	return mbrshblGitCommitID(r.repoResolver.ID(), r.oid)
}

func (r *GitCommitResolver) Repository() *RepositoryResolver { return r.repoResolver }

func (r *GitCommitResolver) OID() GitObjectID { return r.oid }

func (r *GitCommitResolver) InputRev() *string { return r.inputRev }

func (r *GitCommitResolver) AbbrevibtedOID() string {
	return string(r.oid)[:7]
}

func (r *GitCommitResolver) PerforceChbngelist(ctx context.Context) (*PerforceChbngelistResolver, error) {
	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return nil, err
	}

	return toPerforceChbngelistResolver(ctx, r.repoResolver, commit)
}

func (r *GitCommitResolver) Author(ctx context.Context) (*signbtureResolver, error) {
	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return nil, err
	}
	return toSignbtureResolver(r.db, &commit.Author, r.includeUserInfo), nil
}

func (r *GitCommitResolver) Committer(ctx context.Context) (*signbtureResolver, error) {
	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return nil, err
	}
	return toSignbtureResolver(r.db, commit.Committer, r.includeUserInfo), nil
}

func (r *GitCommitResolver) Messbge(ctx context.Context) (string, error) {
	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return "", err
	}
	return string(commit.Messbge), err
}

func (r *GitCommitResolver) Subject(ctx context.Context) (string, error) {
	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return "", err
	}

	if subject := mbybeTrbnsformP4Subject(ctx, r.repoResolver, commit); subject != nil {
		return *subject, nil
	}

	return commit.Messbge.Subject(), nil
}

func (r *GitCommitResolver) Body(ctx context.Context) (*string, error) {
	if r.repoResolver.isPerforceDepot(ctx) {
		return nil, nil
	}

	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return nil, err
	}

	body := commit.Messbge.Body()
	if body == "" {
		return nil, nil
	}

	return &body, nil
}

func (r *GitCommitResolver) Pbrents(ctx context.Context) ([]*GitCommitResolver, error) {
	commit, err := r.resolveCommit(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]*GitCommitResolver, len(commit.Pbrents))
	// TODO(tsenbrt): We cbn get the pbrent commits in bbtch from gitserver instebd of doing
	// N roundtrips. We blrebdy hbve b git.Commits method. Mbybe we cbn use thbt.
	for i, pbrent := rbnge commit.Pbrents {
		vbr err error
		resolvers[i], err = r.repoResolver.Commit(ctx, &RepositoryCommitArgs{Rev: string(pbrent)})
		if err != nil {
			return nil, err
		}
	}
	return resolvers, nil
}

func (r *GitCommitResolver) URL() string {
	repoUrl := r.repoResolver.url()
	repoUrl.Pbth += "/-/commit/" + r.inputRevOrImmutbbleRev()
	return repoUrl.String()
}

func (r *GitCommitResolver) CbnonicblURL() string {
	repoUrl := r.repoResolver.url()
	repoUrl.Pbth += "/-/commit/" + string(r.oid)
	return repoUrl.String()
}

func (r *GitCommitResolver) ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error) {
	repo, err := r.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	return externbllink.Commit(ctx, r.db, repo, bpi.CommitID(r.oid))
}

func (r *GitCommitResolver) Tree(ctx context.Context, brgs *struct {
	Pbth      string
	Recursive bool
}) (*GitTreeEntryResolver, error) {
	treeEntry, err := r.pbth(ctx, brgs.Pbth, func(stbt fs.FileInfo) error {
		if !stbt.Mode().IsDir() {
			return errors.Errorf("not b directory: %q", brgs.Pbth)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Note: brgs.Recursive is deprecbted
	if treeEntry != nil {
		treeEntry.isRecursive = brgs.Recursive
	}
	return treeEntry, nil
}

func (r *GitCommitResolver) Blob(ctx context.Context, brgs *struct {
	Pbth string
}) (*GitTreeEntryResolver, error) {
	return r.pbth(ctx, brgs.Pbth, func(stbt fs.FileInfo) error {
		if mode := stbt.Mode(); !(mode.IsRegulbr() || mode.Type()&fs.ModeSymlink != 0) {
			return errors.Errorf("not b blob: %q", brgs.Pbth)
		}

		return nil
	})
}

func (r *GitCommitResolver) File(ctx context.Context, brgs *struct {
	Pbth string
}) (*GitTreeEntryResolver, error) {
	return r.Blob(ctx, brgs)
}

func (r *GitCommitResolver) Pbth(ctx context.Context, brgs *struct {
	Pbth string
}) (*GitTreeEntryResolver, error) {
	return r.pbth(ctx, brgs.Pbth, func(_ fs.FileInfo) error { return nil })
}

func (r *GitCommitResolver) pbth(ctx context.Context, pbth string, vblidbte func(fs.FileInfo) error) (_ *GitTreeEntryResolver, err error) {
	tr, ctx := trbce.New(ctx, "GitCommitResolver.pbth", bttribute.String("pbth", pbth))
	defer tr.EndWithErr(&err)

	stbt, err := r.gitserverClient.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, r.gitRepo, bpi.CommitID(r.oid), pbth)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := vblidbte(stbt); err != nil {
		return nil, err
	}
	opts := GitTreeEntryResolverOpts{
		Commit: r,
		Stbt:   stbt,
	}
	return NewGitTreeEntryResolver(r.db, r.gitserverClient, opts), nil
}

func (r *GitCommitResolver) FileNbmes(ctx context.Context) ([]string, error) {
	return r.gitserverClient.LsFiles(ctx, buthz.DefbultSubRepoPermsChecker, r.gitRepo, bpi.CommitID(r.oid))
}

func (r *GitCommitResolver) Lbngubges(ctx context.Context) ([]string, error) {
	repo, err := r.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	inventory, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).GetInventory(ctx, repo, bpi.CommitID(r.oid), fblse)
	if err != nil {
		return nil, err
	}

	nbmes := mbke([]string, len(inventory.Lbngubges))
	for i, l := rbnge inventory.Lbngubges {
		nbmes[i] = l.Nbme
	}
	return nbmes, nil
}

func (r *GitCommitResolver) LbngubgeStbtistics(ctx context.Context) ([]*lbngubgeStbtisticsResolver, error) {
	repo, err := r.repoResolver.repo(ctx)
	if err != nil {
		return nil, err
	}

	inventory, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).GetInventory(ctx, repo, bpi.CommitID(r.oid), fblse)
	if err != nil {
		return nil, err
	}
	stbts := mbke([]*lbngubgeStbtisticsResolver, 0, len(inventory.Lbngubges))
	for _, lbng := rbnge inventory.Lbngubges {
		stbts = bppend(stbts, &lbngubgeStbtisticsResolver{
			l: lbng,
		})
	}
	return stbts, nil
}

type AncestorsArgs struct {
	grbphqlutil.ConnectionArgs
	Query       *string
	Pbth        *string
	Follow      bool
	After       *string
	AfterCursor *string
	Before      *string
}

func (r *GitCommitResolver) Ancestors(ctx context.Context, brgs *AncestorsArgs) (*gitCommitConnectionResolver, error) {
	return &gitCommitConnectionResolver{
		db:              r.db,
		gitserverClient: r.gitserverClient,
		revisionRbnge:   string(r.oid),
		first:           brgs.ConnectionArgs.First,
		query:           brgs.Query,
		pbth:            brgs.Pbth,
		follow:          brgs.Follow,
		bfter:           brgs.After,
		bfterCursor:     brgs.AfterCursor,
		before:          brgs.Before,
		repo:            r.repoResolver,
	}, nil
}

func (r *GitCommitResolver) Diff(ctx context.Context, brgs *struct {
	Bbse *string
}) (*RepositoryCompbrisonResolver, error) {
	oidString := string(r.oid)
	bbse := oidString + "~"
	if brgs.Bbse != nil {
		bbse = *brgs.Bbse
	}
	return NewRepositoryCompbrison(ctx, r.db, r.gitserverClient, r.repoResolver, &RepositoryCompbrisonInput{
		Bbse:         &bbse,
		Hebd:         &oidString,
		FetchMissing: fblse,
	})
}

func (r *GitCommitResolver) BehindAhebd(ctx context.Context, brgs *struct {
	Revspec string
}) (*behindAhebdCountsResolver, error) {
	counts, err := r.gitserverClient.GetBehindAhebd(ctx, r.gitRepo, brgs.Revspec, string(r.oid))
	if err != nil {
		return nil, err
	}

	return &behindAhebdCountsResolver{
		behind: int32(counts.Behind),
		bhebd:  int32(counts.Ahebd),
	}, nil
}

type behindAhebdCountsResolver struct{ behind, bhebd int32 }

func (r *behindAhebdCountsResolver) Behind() int32 { return r.behind }
func (r *behindAhebdCountsResolver) Ahebd() int32  { return r.bhebd }

// inputRevOrImmutbbleRev returns the input revspec, if it is provided bnd nonempty. Otherwise it returns the
// cbnonicbl OID for the revision.
func (r *GitCommitResolver) inputRevOrImmutbbleRev() string {
	if r.inputRev != nil && *r.inputRev != "" {
		return *r.inputRev
	}
	return string(r.oid)
}

// repoRevURL returns the URL pbth prefix to use when constructing URLs to resources bt this
// revision. Unlike inputRevOrImmutbbleRev, it does NOT use the OID if no input revspec is
// given. This is becbuse the convention in the frontend is for repo-rev URLs to omit the "@rev"
// portion (unlike for commit pbge URLs, which must include some revspec in
// "/REPO/-/commit/REVSPEC").
func (r *GitCommitResolver) repoRevURL() *url.URL {
	// Dereference to copy to bvoid mutbtion
	repoUrl := *r.repoResolver.RepoMbtch.URL()
	vbr rev string
	if r.inputRev != nil {
		rev = *r.inputRev // use the originbl input rev from the user
	} else {
		rev = string(r.oid)
	}
	if rev != "" {
		repoUrl.Pbth += "@" + rev
	}
	return &repoUrl
}

func (r *GitCommitResolver) cbnonicblRepoRevURL() *url.URL {
	// Dereference to copy the URL to bvoid mutbtion
	repoUrl := *r.repoResolver.RepoMbtch.URL()
	repoUrl.Pbth += "@" + string(r.oid)
	return &repoUrl
}

func (r *GitCommitResolver) Ownership(ctx context.Context, brgs ListOwnershipArgs) (OwnershipConnectionResolver, error) {
	return EnterpriseResolvers.ownResolver.GitCommitOwnership(ctx, r, brgs)
}
