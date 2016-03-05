package local

import (
	"encoding/json"
	"log"
	pathpkg "path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rogpeppe/rog-go/parallel"

	"strings"

	"sort"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/doc"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/pkg/vfsutil"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/repoupdater"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
)

var Repos sourcegraph.ReposServer = &repos{}

var errEmptyRepoURI = grpc.Errorf(codes.InvalidArgument, "repo URI is empty")
var errPermissionDenied = grpc.Errorf(codes.PermissionDenied, "cannot view this repo")

type repos struct{}

var _ sourcegraph.ReposServer = (*repos)(nil)

func (s *repos) Get(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
	if repoSpec.URI == "" {
		return nil, errEmptyRepoURI
	}

	repo, err := store.ReposFromContext(ctx).Get(ctx, repoSpec.URI)
	if err != nil {
		return nil, err
	}

	// Query the remote server for the remote repo, to ensure the
	// actor can access this repo.
	if err := s.setRepoFieldsFromRemote(ctx, repo); err != nil {
		return nil, err
	}

	if repo.Blocked {
		return nil, grpc.Errorf(codes.FailedPrecondition, "repo %s is blocked", repo)
	}
	return repo, nil
}

func (s *repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
	repos, err := store.ReposFromContext(ctx).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	for _, repo_ := range repos {
		repo := repo_
		par.Do(func() error {
			return s.setRepoFieldsFromRemote(ctx, repo)
		})
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}
	return &sourcegraph.RepoList{Repos: repos}, nil
}

func (s *repos) setRepoFieldsFromRemote(ctx context.Context, repo *sourcegraph.Repo) error {
	repo.HTMLURL = conf.AppURL(ctx).ResolveReference(app_router.Rel.URLToRepo(repo.URI)).String()

	// Fetch latest metadata from GitHub (we don't even try to keep
	// our cache up to date).
	if strings.HasPrefix(repo.URI, "github.com/") {
		ghrepo, err := (&github.Repos{}).Get(ctx, repo.URI)
		if err != nil {
			return err
		}
		repo.Description = ghrepo.Description
		repo.Language = ghrepo.Language
		repo.DefaultBranch = ghrepo.DefaultBranch
		repo.Fork = ghrepo.Fork
		repo.Private = ghrepo.Private
		repo.Permissions = ghrepo.Permissions
		repo.UpdatedAt = ghrepo.UpdatedAt
		repo.GitHub = &sourcegraph.GitHubRepo{Stars: ghrepo.Stars}
	}

	return nil
}

func (s *repos) Create(ctx context.Context, op *sourcegraph.ReposCreateOp) (repo *sourcegraph.Repo, err error) {
	switch {
	case op.GetNew() != nil:
		repo, err = s.newRepo(ctx, op.GetNew())
	case op.GetFromGitHubID() != 0:
		repo, err = s.newRepoFromGitHubID(ctx, int(op.GetFromGitHubID()))
	default:
		return nil, grpc.Errorf(codes.Unimplemented, "repo creation operation not supported")
	}

	if err != nil {
		return
	}

	if err := store.ReposFromContext(ctx).Create(ctx, repo); err != nil {
		return nil, err
	}

	repo, err = store.ReposFromContext(ctx).Get(ctx, repo.URI)
	if err != nil {
		return
	}

	if repo.Mirror {
		repoupdater.Enqueue(repo)
	}

	eventsutil.LogAddRepo(ctx, repo.HTTPCloneURL, repo.Language, repo.Mirror, repo.Private)

	return
}

func (s *repos) newRepo(ctx context.Context, op *sourcegraph.ReposCreateOp_NewRepo) (*sourcegraph.Repo, error) {
	if op.URI == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "repo URI must have at least one path component")
	}
	if op.Mirror {
		if op.CloneURL == "" {
			return nil, grpc.Errorf(codes.InvalidArgument, "creating a mirror repo requires a clone URL to be set")
		}
	}

	if authutil.ActiveFlags.PrivateMirrors {
		// TODO: enable creating local repos on Sourcegraph Server instances.
		return nil, grpc.Errorf(codes.Unavailable, "this server only supports creating mirror repos of GitHub repos")
	}

	if op.DefaultBranch == "" {
		op.DefaultBranch = "master"
	}

	ts := pbtypes.NewTimestamp(time.Now())
	return &sourcegraph.Repo{
		Name:          pathpkg.Base(op.URI),
		URI:           op.URI,
		VCS:           op.VCS,
		HTTPCloneURL:  op.CloneURL,
		Language:      op.Language,
		DefaultBranch: op.DefaultBranch,
		Description:   op.Description,
		Mirror:        op.Mirror,
		CreatedAt:     &ts,
	}, nil
}

func (s *repos) newRepoFromGitHubID(ctx context.Context, githubID int) (*sourcegraph.Repo, error) {
	ghrepo, err := (&github.Repos{}).GetByID(ctx, githubID)
	if err != nil {
		return nil, err
	}

	// If this server has a waitlist in place, check that the user
	// is off the waitlist.
	if ghrepo.Private && !authpkg.ActorFromContext(ctx).PrivateMirrors {
		return nil, grpc.Errorf(codes.PermissionDenied, "user is not allowed to create this repo")
	}

	// Purposefully set very few fields. We don't want to cache
	// metadata, because it'll get stale, and fetching online from
	// GitHub is quite easy and (with HTTP caching) performant.
	ts := pbtypes.NewTimestamp(time.Now())
	return &sourcegraph.Repo{
		Name:         ghrepo.Name,
		URI:          "github.com/" + ghrepo.Owner + "/" + ghrepo.Name,
		HTTPCloneURL: ghrepo.HTTPCloneURL,
		VCS:          "git",
		Mirror:       true,
		CreatedAt:    &ts,
	}, nil
}

func (s *repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (*sourcegraph.Repo, error) {
	ts := time.Now()
	update := &store.RepoUpdate{ReposUpdateOp: op, UpdatedAt: &ts}
	if err := store.ReposFromContext(ctx).Update(ctx, update); err != nil {
		return nil, err
	}
	return s.Get(ctx, &op.Repo)
}

func (s *repos) Delete(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	if err := store.ReposFromContext(ctx).Delete(ctx, repo.URI); err != nil {
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
			// TODO(sqs): check this
			repoRev.CommitID = dataVer.CommitID
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
	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{URI: repoURI})
	if err != nil {
		return "", err
	}
	if repo.DefaultBranch == "" {
		return "", grpc.Errorf(codes.FailedPrecondition, "repo %s has no default branch", repoURI)
	}
	return repo.DefaultBranch, nil
}

func (s *repos) GetReadme(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*sourcegraph.Readme, error) {
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

	commit := vcs.CommitID(repoRev.CommitID)

	entries, err := vcsrepo.ReadDir(commit, ".", false)
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

	data, err := vcsrepo.ReadFile(commit, readme.Path)
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
	repoConfigsStore := store.RepoConfigsFromContext(ctx)

	conf, err := repoConfigsStore.Get(ctx, repo.URI)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &sourcegraph.RepoConfig{}
	}
	return conf, nil
}

func (s *repos) ConfigureApp(ctx context.Context, op *sourcegraph.RepoConfigureAppOp) (*pbtypes.Void, error) {
	store := store.RepoConfigsFromContext(ctx)

	if op.Enable {
		// Check that app ID is a valid app. Allow disabling invalid
		// apps so that obsolete apps can always be removed.
		if _, present := platform.Frames()[op.App]; !present {
			return nil, grpc.Errorf(codes.InvalidArgument, "app %q is not a valid app ID", op.App)
		}
	}

	conf, err := store.Get(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &sourcegraph.RepoConfig{}
	}

	// Make apps list unique and add/remove the new app.
	apps := make(map[string]struct{}, len(conf.Apps))
	for _, app := range conf.Apps {
		apps[app] = struct{}{}
	}
	if op.Enable {
		apps[op.App] = struct{}{}
	} else {
		delete(apps, op.App)
	}
	conf.Apps = make([]string, 0, len(apps))
	for app := range apps {
		conf.Apps = append(conf.Apps, app)
	}
	sort.Strings(conf.Apps)

	if err := store.Update(ctx, op.Repo.URI, *conf); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *repos) GetInventory(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	if localcli.Flags.DisableRepoInventory {
		return nil, grpc.Errorf(codes.Unimplemented, "repo inventory listing is disabled by the configuration (DisableRepoInventory/--local.disable-repo-inventory)")
	}

	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
		return nil, err
	}

	// Consult the commit status "cache" for a cached inventory result.
	//
	// We cache inventory result on the commit status. This lets us
	// populate the cache by calling this method from anywhere (e.g.,
	// after a git push). Just using the memory cache would mean that
	// each server process would have to recompute this result.
	const statusContext = "cache:repo.inventory"
	statusRev := sourcegraph.RepoRevSpec{RepoSpec: repoRev.RepoSpec, CommitID: repoRev.CommitID}
	statuses, err := svc.RepoStatuses(ctx).GetCombined(ctx, &statusRev)
	if err != nil {
		return nil, err
	}
	if status := statuses.GetStatus(statusContext); status != nil {
		var inv inventory.Inventory
		if err := json.Unmarshal([]byte(status.Description), &inv); err == nil {
			return &inv, nil
		}
		log15.Warn("Repos.GetInventory failed to unmarshal cached JSON inventory", "repoRev", statusRev, "err", err)
	}

	// Not found in the cache, so compute it.
	inv, err := s.getInventoryUncached(ctx, repoRev)
	if err != nil {
		return nil, err
	}

	// Store inventory in cache.
	jsonData, err := json.Marshal(inv)
	if err != nil {
		return nil, err
	}

	_, err = svc.RepoStatuses(ctx).Create(ctx, &sourcegraph.RepoStatusesCreateOp{
		Repo:   statusRev,
		Status: sourcegraph.RepoStatus{Description: string(jsonData), Context: statusContext},
	})
	if err != nil {
		log15.Warn("Failed to update RepoStatuses cache", "err", err, "Repo URI", repoRev.RepoSpec.URI)
	}

	return inv, nil
}

func (s *repos) getInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.URI)
	if err != nil {
		return nil, err
	}

	fs := vcs.FileSystem(vcsrepo, vcs.CommitID(repoRev.CommitID))
	inv, err := inventory.Scan(ctx, vfsutil.Walkable(fs, filepath.Join))
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *repos) verifyScopeHasPrivateRepoAccess(scope map[string]bool) bool {
	for k := range scope {
		if strings.HasPrefix(k, "internal:") {
			return true
		}
	}
	return false
}
