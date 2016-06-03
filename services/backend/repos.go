package backend

import (
	"encoding/json"
	"fmt"
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
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vfsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	localcli "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/services/platform"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/test/e2e/e2etestuser"
	"sourcegraph.com/sqs/pbtypes"
)

var Repos sourcegraph.ReposServer = &repos{}

var errEmptyRepoURI = grpc.Errorf(codes.InvalidArgument, "repo URI is empty")

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

	// HACK(beyang): this behavior should really be set by an options struct
	// field and not depend on permissions grants. Callers of this method
	// assume DefaultBranch will be set right now, but it is not always set.
	// For now, always set remote fields if the repository is public. (It
	// cannot be set for private repositories, because otherwise the import
	// step of builds will break).
	//
	// If the actor doesn't have a special grant to access this repo,
	// query the remote server for the remote repo, to ensure the
	// actor can access this repo.
	//
	// Special grants are given to drone workers to fetch repo metadata
	// when configuring a build.
	hasGrant := accesscontrol.VerifyScopeHasAccess(ctx, authpkg.ActorFromContext(ctx).Scope, "Repos.Get", repoSpec.URI)
	if !hasGrant {
		if err := s.setRepoFieldsFromRemote(ctx, repo); err != nil {
			return nil, err
		}
	} else {
		// if the actor does have a special grant (e.g., a worker), still best-effort attempt to set fields from remote.
		s.setRepoFieldsFromRemote(ctx, repo)
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
	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		ghrepo, err := (&github.Repos{}).Get(ctx, repo.URI)
		if err != nil {
			return err
		}
		repoSetFromRemote(repo, ghrepo)
	}

	return nil
}

func repoSetFromRemote(repo *sourcegraph.Repo, ghrepo *sourcegraph.RemoteRepo) {
	repo.Description = ghrepo.Description
	repo.Language = ghrepo.Language
	repo.DefaultBranch = ghrepo.DefaultBranch
	repo.Fork = ghrepo.Fork
	repo.Private = ghrepo.Private
	repo.UpdatedAt = ghrepo.UpdatedAt
	repo.PushedAt = ghrepo.PushedAt
}

func (s *repos) Create(ctx context.Context, op *sourcegraph.ReposCreateOp) (repo *sourcegraph.Repo, err error) {
	switch {
	case op.GetNew() != nil:
		repo, err = s.newRepo(ctx, op.GetNew())
		if err != nil {
			return
		}
	case op.GetFromGitHubID() != 0:
		repo, err = s.newRepoFromGitHubID(ctx, int(op.GetFromGitHubID()))
		if err != nil {
			return
		}
	default:
		return nil, grpc.Errorf(codes.Unimplemented, "repo creation operation not supported")
	}

	if err := store.ReposFromContext(ctx).Create(ctx, repo); err != nil {
		return nil, err
	}

	repo, err = s.Get(ctx, &sourcegraph.RepoSpec{URI: repo.URI})
	if err != nil {
		return
	}

	if repo.Mirror {
		var asUser *sourcegraph.UserSpec
		if actor := authpkg.ActorFromContext(ctx); actor.UID != 0 {
			asUser = &sourcegraph.UserSpec{UID: int32(actor.UID), Login: actor.Login}
		}
		repoupdater.Enqueue(repo.RepoSpec(), asUser)
	}

	sendCreateRepoSlackMsg(ctx, repo.URI, repo.Language, repo.Mirror, repo.Private)

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
	if strings.HasPrefix(strings.ToLower(op.URI), "github.com/") {
		if !op.Mirror {
			return nil, grpc.Errorf(codes.InvalidArgument, "github.com/ repos can only be mirrors")
		}
		if !(op.CloneURL == fmt.Sprintf("https://%s", op.URI) || op.CloneURL == fmt.Sprintf("https://%s.git", op.URI)) {
			// Disallow creating GitHub mirrors via repo URI and clone URL unless they match.
			return nil, grpc.Errorf(codes.InvalidArgument, "github.com/ mirrors repos can only be created if the clone URL matches the repo URI")
		}
	}

	if op.DefaultBranch == "" {
		op.DefaultBranch = "master"
	}

	ts := pbtypes.NewTimestamp(time.Now())
	return &sourcegraph.Repo{
		Name:          pathpkg.Base(op.URI),
		URI:           op.URI,
		HTTPCloneURL:  op.CloneURL,
		Language:      op.Language,
		DefaultBranch: op.DefaultBranch,
		Description:   op.Description,
		Mirror:        op.Mirror,
		CreatedAt:     &ts,
	}, nil
}

func (s *repos) newRepoFromGitHubID(ctx context.Context, gitHubID int) (*sourcegraph.Repo, error) {
	ghrepo, err := (&github.Repos{}).GetByID(ctx, gitHubID)
	if err != nil {
		return nil, err
	}

	// Purposefully set very few fields. We don't want to cache
	// metadata, because it'll get stale, and fetching online from
	// GitHub is quite easy and (with HTTP caching) performant.
	ts := pbtypes.NewTimestamp(time.Now())
	return &sourcegraph.Repo{
		Owner:        ghrepo.Owner,
		Name:         ghrepo.Name,
		URI:          githubutil.RepoURI(ghrepo.Owner, ghrepo.Name),
		HTTPCloneURL: ghrepo.HTTPCloneURL,
		Description:  ghrepo.Description,
		Mirror:       true,
		Fork:         ghrepo.Fork,
		CreatedAt:    &ts,

		// KLUDGE: set this to be true to avoid accidentally treating
		// a private GitHub repo as public (the real value should be
		// populated from GitHub on the fly).
		Private: true,
	}, nil
}

func (s *repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (*sourcegraph.Repo, error) {
	ts := time.Now()
	update := store.RepoUpdate{ReposUpdateOp: op, UpdatedAt: &ts}
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
		if _, present := platform.Apps[op.App]; !present {
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

	if !isAbsCommitID(repoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	// Consult the commit status "cache" for a cached inventory result.
	//
	// We cache inventory result on the commit status. This lets us
	// populate the cache by calling this method from anywhere (e.g.,
	// after a git push). Just using the memory cache would mean that
	// each server process would have to recompute this result.
	const statusContext = "cache:repo.inventory"
	statusRev := sourcegraph.RepoRevSpec{Repo: repoRev.Repo, CommitID: repoRev.CommitID}
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
		log15.Warn("Failed to update RepoStatuses cache", "err", err, "Repo URI", repoRev.Repo)
	}

	return inv, nil
}

func (s *repos) getInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.Repo)
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

func sendCreateRepoSlackMsg(ctx context.Context, uri, language string, mirror, private bool) {
	user := authpkg.ActorFromContext(ctx).Login
	if strings.HasPrefix(user, e2etestuser.Prefix) {
		return
	}

	repoType := "public"
	if private {
		repoType = "private"
	}
	if mirror {
		repoType += " mirror"
	} else {
		repoType += " hosted"
	}

	msg := fmt.Sprintf("User *%s* added a %s repo", user, repoType)
	if !private {
		msg += fmt.Sprintf(": <https://sourcegraph.com/%s|%s>", uri, uri)
	}
	if language != "" {
		msg += fmt.Sprintf(" (%s)", language)
	}
	notif.PostOnboardingNotif(msg)
}
