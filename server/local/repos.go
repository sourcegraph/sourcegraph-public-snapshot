package local

import (
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/doc"
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
	if repoRev.Resolved() {
		return nil
	}

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

func (s *repos) Enable(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	return &pbtypes.Void{}, s.enableOrDisable(ctx, *repo, true)
}

func (s *repos) Disable(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	return &pbtypes.Void{}, s.enableOrDisable(ctx, *repo, false)
}

// enableOrDisable updates a repo's config to be either enabled or
// disabled. It also triggers hooks that must run after the config
// changes (and it diffs the new settings from the current to only
// trigger hooks when a config property actually changes).
func (s *repos) enableOrDisable(ctx context.Context, repo sourcegraph.RepoSpec, enable bool) error {
	defer noCache(ctx)

	// Get current config.
	cur, err := svc.Repos(ctx).GetConfig(ctx, &repo)
	if err != nil {
		return err
	}

	store := store.RepoConfigsFromContextOrNil(ctx)
	if store == nil {
		return &sourcegraph.NotImplementedError{What: "RepoConfigs"}
	}

	// We might need to fetch the repo to perform the config update.
	var r *sourcegraph.Repo

	// Compute the updated config.
	updated := *cur
	updated.Enabled = enable
	if a := authpkg.ActorFromContext(ctx); a.IsAuthenticated() {
		cur.LastAdminUID = int32(a.UID)
	}

	// Execute hooks before updating the config, in case a hook fails
	// (so the config doesn't erroneously reflect the desired final
	// state).
	if updated.Enabled != cur.Enabled {
		// Fetch repo if haven't already.
		if r == nil {
			var err error
			r, err = s.get(ctx, repo.URI)
			if err != nil {
				return err
			}
		}
		if err := s.changedConfig_Enabled(ctx, r, updated.Enabled); err != nil {
			return err
		}
	}

	return store.Update(ctx, repo.URI, updated)
}

func (s *repos) GetConfig(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
	repoConfigsStore := store.RepoConfigsFromContextOrNil(ctx)
	if repoConfigsStore == nil {
		// Reasonable default.
		return &sourcegraph.RepoConfig{Enabled: true}, nil
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
