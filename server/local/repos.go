package local

import (
	"io/ioutil"
	"log"
	"time"

	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/doc"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var Repos sourcegraph.ReposServer = &repos{}

var errEmptyRepoURI = &sourcegraph.InvalidSpecError{Reason: "repo URI is empty"}

type repos struct{}

var _ sourcegraph.ReposServer = (*repos)(nil)

func (s *repos) Get(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
	r, err := s.get(ctx, repo.URI)
	if err != nil {
		return nil, err
	}
	if err := s.setRepoPermissions(ctx, r); err != nil {
		return nil, err
	}
	shortCache(ctx)
	return r, nil
}

// get gets the repo from the store but does not fetch and populate
// the repo permissions. Callers that need the repo but not the
// permissions should call get (instead of Get) to avoid doing
// needless work.
func (s *repos) get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	if repo == "" {
		return nil, errEmptyRepoURI
	}

	r, err := store.ReposFromContext(ctx).Get(ctx, repo)
	if err != nil {
		return nil, err
	}

	if r.Blocked {
		return nil, grpc.Errorf(codes.FailedPrecondition, "repo %s is blocked", repo)
	}

	return r, nil
}

func (s *repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
	repos, err := store.ReposFromContext(ctx).List(ctx, opt)
	if err != nil {
		return nil, err
	}
	if err := s.setRepoPermissions(ctx, repos...); err != nil {
		return nil, err
	}
	veryShortCache(ctx)
	return &sourcegraph.RepoList{Repos: repos}, nil
}

// setRepoPermissions modifies repos in place, setting their
// Permissions fields by calling (store.Repos).GetPerms on each repo.
func (s *repos) setRepoPermissions(ctx context.Context, repos ...*sourcegraph.Repo) error {
	for _, repo := range repos {
		if repo.Permissions == nil {
			perms, err := store.ReposFromContext(ctx).GetPerms(ctx, repo.URI)
			if err != nil {
				return err
			}
			repo.Permissions = perms
		}
	}
	return nil
}

func (s *repos) Create(ctx context.Context, op *sourcegraph.ReposCreateOp) (*sourcegraph.Repo, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Create"); err != nil {
		return nil, err
	}

	ts := pbtypes.NewTimestamp(time.Now())
	repo := &sourcegraph.Repo{
		URI:          op.URI,
		VCS:          op.VCS,
		HTTPCloneURL: op.CloneURL,
		Mirror:       op.Mirror,
		Private:      op.Private,
		Description:  op.Description,
		Language:     op.Language,
		CreatedAt:    &ts,
	}
	if _, err := store.ReposFromContextOrNil(ctx).Create(ctx, repo); err != nil {
		return nil, err
	}
	return store.ReposFromContext(ctx).Get(ctx, op.URI)
}

func (s *repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (*sourcegraph.Repo, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Update"); err != nil {
		return nil, err
	}

	ts := time.Now()
	update := &store.RepoUpdate{ReposUpdateOp: op, UpdatedAt: &ts}
	if err := store.ReposFromContext(ctx).Update(ctx, update); err != nil {
		return nil, err
	}
	return s.get(ctx, op.Repo.URI)
}

func (s *repos) Delete(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Delete"); err != nil {
		return nil, err
	}

	if err := store.ReposFromContextOrNil(ctx).Delete(ctx, repo.URI); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

// resolveRepoRev resolves repoRev to an absolute commit ID (by
// consulting its VCS data). If no rev is specified, the repo's
// default branch is used.
func (s *repos) resolveRepoRev(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) error {
	if err := s.resolveRepoRevBranch(ctx, repoRev); err != nil {
		return err
	}

	if repoRev.CommitID == "" {
		vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.URI)
		if err != nil {
			return err
		}
		commitID, err := vcsrepo.ResolveRevision(repoRev.Rev)
		if err != nil {
			return err
		}
		repoRev.CommitID = string(commitID)
	}

	return nil
}

func (s *repos) resolveRepoRevBranch(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) error {
	if repoRev.CommitID == "" && repoRev.Rev == "" {
		// Get default branch.
		defBr, err := s.defaultBranch(ctx, repoRev.URI)
		if err != nil {
			return err
		}
		repoRev.Rev = defBr
	}

	const srclibRevTag = "^{srclib}" // REV^{srclib} refers to the newest srclib version from REV
	if strings.HasSuffix(repoRev.Rev, srclibRevTag) {
		origRev := repoRev.Rev
		repoRev.Rev = strings.TrimSuffix(repoRev.Rev, srclibRevTag)
		dataVer, err := svc.Repos(ctx).GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{RepoRev: *repoRev})
		if err == nil {
			repoRev.Rev = dataVer.CommitID
		} else if errcode.GRPC(err) == codes.NotFound {
			// Ignore NotFound as otherwise the user might not even be
			// able to access the repository homepage.
			log15.Warn("Failed to resolve to commit with srclib Code Intelligence data; will proceed by resolving to commit with no Code Intelligence data instead", "rev", origRev, "fallback", repoRev.Rev, "error", err)
		} else if err != nil {
			return grpc.Errorf(errcode.GRPC(err), "while resolving rev %q: %s", repoRev.Rev, err)
		}
	}

	return nil
}

func (s *repos) defaultBranch(ctx context.Context, repoURI string) (string, error) {
	repo, err := s.get(ctx, repoURI)
	if err != nil {
		return "", err
	}
	if repo.DefaultBranch == "" {
		return "", grpc.Errorf(codes.FailedPrecondition, "repo %s has no default branch", repoURI)
	}
	return repo.DefaultBranch, nil
}

func (s *repos) GetReadme(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*sourcegraph.Readme, error) {
	cacheOnCommitID(ctx, repoRev.CommitID)

	if repoRev.URI == "" {
		return nil, errEmptyRepoURI
	}

	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.URI)
	if err != nil {
		return nil, err
	}
	fs, err := vcsrepo.FileSystem(vcs.CommitID(repoRev.CommitID))
	if err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	filenames := make([]string, len(entries))
	for i, e := range entries {
		filenames[i] = e.Name()
	}

	readme := &sourcegraph.Readme{Path: doc.ChooseReadme(filenames)}
	if readme.Path == "" {
		return nil, grpc.Errorf(codes.NotFound, "no README found in %v", repoRev)
	}

	f, err := fs.Open(readme.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	formatted, err := doc.ToHTML(doc.Format(readme.Path), data)
	if err != nil {
		log.Printf("Warning: doc.ToHTML on readme %q in repo %s failed: %s.", readme.Path, repoRev.URI, err)
	}
	readme.HTML = string(formatted)
	return readme, nil
}

func (s *repos) GetConfig(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
	repoConfigsStore := store.RepoConfigsFromContextOrNil(ctx)
	if repoConfigsStore == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "RepoConfigs is not implemented")
	}

	conf, err := repoConfigsStore.Get(ctx, repo.URI)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &sourcegraph.RepoConfig{}
	}
	return conf, nil
}
