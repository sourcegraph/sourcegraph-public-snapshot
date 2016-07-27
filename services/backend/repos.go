package backend

import (
	"encoding/json"
	"fmt"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/neelance/parallel"

	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vfsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	localcli "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/test/e2e/e2etestuser"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sqs/pbtypes"
)

var Repos sourcegraph.ReposServer = &repos{}

type repos struct{}

var _ sourcegraph.ReposServer = (*repos)(nil)

func (s *repos) Get(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
	repo, err := store.ReposFromContext(ctx).Get(ctx, repoSpec.ID)
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
	hasGrant := accesscontrol.VerifyScopeHasAccess(ctx, authpkg.ActorFromContext(ctx).Scope, "Repos.Get", repo.ID)
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

	// deletedRepoURI is used to prune repos from list. Its value should be something an existing repo cannot have, like empty string.
	const deletedRepoURI = ""

	par := parallel.NewRun(30)
	for _, repo_ := range repos {
		repo := repo_
		par.Acquire()
		go func() {
			defer par.Release()
			// TODO(shurcooL): Now that the store is doing more of this, investigate if setRepoFieldsFromRemote
			//                 is still necceessary to do here, or if it can be optimized away.
			if err := s.setRepoFieldsFromRemote(ctx, repo); err != nil {
				if grpc.Code(err) == codes.NotFound {
					// This can happen if a repo is disabled; it will be included by list operation,
					// but getting it directly will result in 404. Treat it as a missing repository,
					// mark it to be removed from the list afterwards.
					log15.Debug("Repo from list not found on remote", "repo", repo.URI)
					repo.URI = deletedRepoURI
					return
				}
				par.Error(err)
				return
			}
		}()
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	// Clear any missing repos with URI == deletedRepoURI after setRepoFieldsFromRemote step.
	for i := 0; i < len(repos); {
		if repos[i].URI == deletedRepoURI {
			// Delete missing repo from list.
			copy(repos[i:], repos[i+1:])
			repos[len(repos)-1] = nil
			repos = repos[:len(repos)-1]
			continue
		}
		i++
	}

	return &sourcegraph.RepoList{Repos: repos}, nil
}

//  ListDeps lists dependencies for a given list of repo URIs.
//
// TODO properly support using repo IDs instead of URIs.
func (s *repos) ListDeps(ctx context.Context, repos *sourcegraph.URIList) (*sourcegraph.URIList, error) {
	repoFilters := []srcstore.UnitFilter{
		srcstore.ByRepos(repos.URIs...),
	}
	units, err := store.GraphFromContext(ctx).Units(repoFilters...)
	if err != nil {
		return nil, err
	}

	deps := make(map[string]struct{})
	for _, u := range units {
		for _, d := range u.Info.Dependencies {
			deps[d.Name] = struct{}{}
		}
	}
	uris := []string{}
	for d, _ := range deps {
		uris = append(uris, d)
	}

	return &sourcegraph.URIList{
		URIs: uris,
	}, nil
}

func (s *repos) setRepoFieldsFromRemote(ctx context.Context, repo *sourcegraph.Repo) error {
	var updateWithDataFromRemote *store.RepoUpdate
	repo.HTMLURL = conf.AppURL(ctx).ResolveReference(app_router.Rel.URLToRepo(repo.URI)).String()
	// Fetch latest metadata from GitHub (we don't even try to keep
	// our cache up to date).
	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		ghrepo, err := github.ReposFromContext(ctx).Get(ctx, repo.URI)
		if err != nil {
			return err
		}
		updateWithDataFromRemote = repoSetFromRemote(repo, ghrepo)
	}

	if updateWithDataFromRemote != nil {
		log15.Debug("Updating repo metadata from remote", "repo", repo.URI)
		// setRepoFieldsFromRemote is used in read requests, including
		// unauthed ones. However, this write isn't as the user, but
		// rather an optimization for us to save the data from
		// github. As such we use an elevated context to allow the
		// write.
		if err := store.ReposFromContext(ctx).Update(elevatedActor(ctx), *updateWithDataFromRemote); err != nil {
			log15.Error("Failed updating repo metadata from remote", "repo", repo.URI, "error", err)
		}
	}

	return nil
}

func timestampEqual(a, b *pbtypes.Timestamp) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Seconds == b.Seconds && a.Nanos == b.Nanos
}

// repoSetFromRemote updates repo with fields from ghrepo that are
// different. If any fields are changed a non-nil store.RepoUpdate is returned
// representing the update.
func repoSetFromRemote(repo *sourcegraph.Repo, ghrepo *sourcegraph.Repo) *store.RepoUpdate {
	updateOp := &store.RepoUpdate{
		ReposUpdateOp: &sourcegraph.ReposUpdateOp{
			Repo: repo.ID,
		},
	}
	updated := false
	if repo.Description != ghrepo.Description {
		repo.Description = ghrepo.Description
		updateOp.Description = ghrepo.Description
		updated = true
	}
	if repo.Language != ghrepo.Language {
		repo.Language = ghrepo.Language
		updateOp.Language = ghrepo.Language
		updated = true
	}
	if repo.DefaultBranch != ghrepo.DefaultBranch {
		repo.DefaultBranch = ghrepo.DefaultBranch
		updateOp.DefaultBranch = ghrepo.DefaultBranch
		updated = true
	}
	if repo.Fork != ghrepo.Fork {
		repo.Fork = ghrepo.Fork
		if ghrepo.Fork {
			updateOp.Fork = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Fork = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if repo.Private != ghrepo.Private {
		repo.Private = ghrepo.Private
		if ghrepo.Private {
			updateOp.Private = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Private = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if !timestampEqual(repo.UpdatedAt, ghrepo.UpdatedAt) {
		repo.UpdatedAt = ghrepo.UpdatedAt
		if ghrepo.UpdatedAt != nil {
			t := ghrepo.UpdatedAt.Time()
			updateOp.UpdatedAt = &t
		}
		updated = true
	}
	if !timestampEqual(repo.PushedAt, ghrepo.PushedAt) {
		repo.PushedAt = ghrepo.PushedAt
		if ghrepo.PushedAt != nil {
			t := ghrepo.PushedAt.Time()
			updateOp.PushedAt = &t
		}
		updated = true
	}

	if updated {
		return updateOp
	}
	return nil
}

func (s *repos) Create(ctx context.Context, op *sourcegraph.ReposCreateOp) (repo *sourcegraph.Repo, err error) {
	switch {
	case op.GetNew() != nil:
		repo, err = s.newRepo(ctx, op.GetNew())
		if err != nil {
			return
		}
	case op.GetFromGitHubID() != 0:
		repo, err = s.newRepoFromOrigin(ctx, &sourcegraph.Origin{
			Service: sourcegraph.Origin_GitHub,
			ID:      strconv.Itoa(int(op.GetFromGitHubID())),
		})
		if err != nil {
			return
		}
	case op.GetOrigin() != nil:
		repo, err = s.newRepoFromOrigin(ctx, op.GetOrigin())
		if err != nil {
			return
		}
	default:
		return nil, grpc.Errorf(codes.Unimplemented, "repo creation operation not supported")
	}

	repoID, err := store.ReposFromContext(ctx).Create(ctx, repo)
	if err != nil {
		return nil, err
	}
	repo.ID = repoID

	repo, err = s.Get(ctx, &sourcegraph.RepoSpec{ID: repo.ID})
	if err != nil {
		return
	}

	repoMaybeEnqueueUpdate(ctx, repo)
	sendCreateRepoSlackMsg(ctx, repo.URI, repo.Language, repo.Mirror, repo.Private)

	return
}

// repoMaybeEnqueueUpdate enqueues an update as the current user if the repo
// is a Mirror.
func repoMaybeEnqueueUpdate(ctx context.Context, repo *sourcegraph.Repo) {
	if !repo.Mirror {
		return
	}
	var asUser *sourcegraph.UserSpec
	if actor := authpkg.ActorFromContext(ctx); actor.UID != 0 {
		asUser = &sourcegraph.UserSpec{UID: int32(actor.UID), Login: actor.Login}
	}
	repoupdater.Enqueue(repo.ID, asUser)
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

func (s *repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (*sourcegraph.Repo, error) {
	ts := time.Now()
	update := store.RepoUpdate{ReposUpdateOp: op, UpdatedAt: &ts}
	if err := store.ReposFromContext(ctx).Update(ctx, update); err != nil {
		return nil, err
	}

	return s.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
}

func (s *repos) Delete(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	if err := store.ReposFromContext(ctx).Delete(ctx, repo.ID); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *repos) GetConfig(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
	conf, err := store.RepoConfigsFromContext(ctx).Get(ctx, repo.ID)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &sourcegraph.RepoConfig{}
	}
	return conf, nil
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
	statuses, err := svc.RepoStatuses(ctx).GetCombined(ctx, repoRev)
	if err != nil {
		return nil, err
	}
	if status := statuses.GetStatus(statusContext); status != nil {
		var inv inventory.Inventory
		if err := json.Unmarshal([]byte(status.Description), &inv); err == nil {
			return &inv, nil
		}
		log15.Warn("Repos.GetInventory failed to unmarshal cached JSON inventory", "repoRev", repoRev, "err", err)
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
		Repo:   *repoRev,
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
