package local

import (
	"encoding/json"
	"io/ioutil"
	"log"
	pathpkg "path"
	"time"

	"strings"

	"sort"

	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/grpccache"
	"sourcegraph.com/sqs/pbtypes"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/doc"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/repoupdater"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
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
	if err := s.setRepoOtherFields(ctx, r); err != nil {
		return nil, err
	}
	veryShortCache(ctx)
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
	if err := s.setRepoOtherFields(ctx, repos...); err != nil {
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

func (s *repos) setRepoOtherFields(ctx context.Context, repos ...*sourcegraph.Repo) error {
	appURL := conf.AppURL(ctx)
	for _, repo := range repos {
		repo.HTMLURL = appURL.ResolveReference(app_router.Rel.URLToRepo(repo.URI)).String()
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
	repo, err := store.ReposFromContextOrNil(ctx).Create(ctx, repo)
	if err != nil {
		return nil, err
	}

	repoupdater.Enqueue(repo)

	repoSpec := repo.RepoSpec()
	return s.Get(ctx, &repoSpec)
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
	return s.Get(ctx, &op.Repo)
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
		vcsrepo, err := cachedRepoVCSOpen(ctx, repoRev.URI)
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

	vcsrepo, err := cachedRepoVCSOpen(ctx, repoRev.URI)
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

func (s *repos) ConfigureApp(ctx context.Context, op *sourcegraph.RepoConfigureAppOp) (*pbtypes.Void, error) {
	defer noCache(ctx)

	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Repos.ConfigureApp"); err != nil {
		return nil, err
	}

	store := store.RepoConfigsFromContextOrNil(ctx)
	if store == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "RepoConfigs is not implemented")
	}

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

	defer cacheOnCommitID(ctx, repoRev.CommitID)

	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
		return nil, err
	}

	// Consult the commit status "cache" for a cached inventory result.
	//
	// We cache inventory result on the commit status. This lets us
	// populate the cache by calling this method from anywhere (e.g.,
	// after a git push). Just using grpccache would only cache the
	// result in that particular client.
	const statusContext = "cache:repo.inventory"
	statusRev := sourcegraph.RepoRevSpec{RepoSpec: repoRev.RepoSpec, CommitID: repoRev.CommitID}
	statuses, err := svc.RepoStatuses(ctx).GetCombined(ctx, &statusRev)
	if err != nil {
		return nil, err
	}
	if status := statuses.GetStatus(statusContext); status != nil {
		var inv inventory.Inventory
		if err := json.Unmarshal([]byte(status.Description), &inv); err == nil {
			if grpccache.GetNoCache(ctx) {
				// Warn because debugging caching issues can be hard.
				log15.Info("Repos.GetInventory is using cached inventory despite gRPC set to no-cache because inventory can be very slow (remove the commit status from the cache if needed)", "repoRev", statusRev)
			}
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
		return nil, err
	}

	return inv, nil
}

func (s *repos) getInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	vcsrepo, err := cachedRepoVCSOpen(ctx, repoRev.URI)
	if err != nil {
		return nil, err
	}

	fs, err := vcsrepo.FileSystem(vcs.CommitID(repoRev.CommitID))
	if err != nil {
		return nil, err
	}

	inv, err := inventory.Scan(ctx, walkableFileSystem{fs})
	if err != nil {
		return nil, err
	}
	return inv, nil
}

type walkableFileSystem struct{ vfs.FileSystem }

func (walkableFileSystem) Join(path ...string) string { return pathpkg.Join(path...) }
