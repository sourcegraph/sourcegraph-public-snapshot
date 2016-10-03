package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	gogithub "github.com/sourcegraph/go-github/github"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vfsutil"
	localcli "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
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
	repo, err := localstore.Repos.Get(ctx, repoSpec.ID)
	if err != nil {
		return nil, err
	}

	if err := s.setRepoFieldsFromRemote(ctx, repo); err != nil {
		return nil, err
	}

	if repo.Blocked {
		return nil, grpc.Errorf(codes.FailedPrecondition, "repo %s is blocked", repo)
	}

	return repo, nil
}

// ghRepoQueryMatcher matches search queries that look like they refer
// to GitHub repositories. Examples include "github.com/gorilla/mux", "gorilla/mux", "gorilla mux",
// "gorilla / mux"
var ghRepoQueryMatcher = regexp.MustCompile(`^(?:github.com/)?([^/\s]+)[/\s]+([^/\s]+)$`)

func (s *repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.List")
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	if opt.RemoteOnly {
		repos, err := github.ListAllGitHubRepos(ctx, &gogithub.RepositoryListOptions{Type: opt.Type})
		if err != nil {
			return nil, err
		}
		return &sourcegraph.RepoList{Repos: repos}, nil
	}

	repos, err := localstore.Repos.List(ctx, &localstore.RepoListOp{
		Name:        opt.Name,
		Query:       opt.Query,
		URIs:        opt.URIs,
		Sort:        opt.Sort,
		Direction:   opt.Direction,
		NoFork:      opt.NoFork,
		Type:        opt.Type,
		Owner:       opt.Owner,
		ListOptions: opt.ListOptions,
	})
	if err != nil {
		return nil, err
	}

	// Augment with external results if user is authenticated,
	// RemoteSearch is true, and Query is non-empty.
	if authpkg.ActorFromContext(ctx).IsAuthenticated() && opt.RemoteSearch {
		ghquery := opt.Query
		if matches := ghRepoQueryMatcher.FindStringSubmatch(opt.Query); matches != nil {
			// Apply query transformation to make GitHub results better.
			ghquery = fmt.Sprintf("user:%s in:name %s", matches[1], matches[2])
		}

		var ghrepos []*sourcegraph.Repo
		var err error
		if ghquery == "" {
			ghrepos, err = github.ListAllGitHubRepos(ctx, &gogithub.RepositoryListOptions{Type: opt.Type})
			ghrepos, repos = repos, ghrepos
		} else {
			ghrepos, err = github.ReposFromContext(ctx).Search(ctx, ghquery, nil)
		}
		if err == nil {
			existingRepos := make(map[string]struct{}, len(repos))
			for _, repo := range repos {
				existingRepos[repo.URI] = struct{}{}
			}
			for _, ghrepo := range ghrepos {
				if _, in := existingRepos[ghrepo.URI]; !in {
					repos = append(repos, ghrepo)
				}
			}
		} else {
			// Fetching results from GitHub is best-effort, as we
			// might hit the rate limit and don't want this to kill
			// the search experience entirely.
			log15.Warn("Unable to fetch repo search results from GitHub", "query", opt.Query, "error", err)
		}
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
	units, err := localstore.Graph.Units(repoFilters...)
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

// setRepoFieldsFromRemote sets the fields of the repository from the
// remote (e.g., GitHub) and updates the repository in the store layer.
func (s *repos) setRepoFieldsFromRemote(ctx context.Context, repo *sourcegraph.Repo) error {
	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		// Fetch latest metadata from GitHub
		ghrepo, err := github.ReposFromContext(ctx).Get(ctx, repo.URI)
		if err != nil {
			return err
		}
		if update := repoSetFromRemote(repo, ghrepo); update != nil {
			log15.Debug("Updating repo metadata from remote", "repo", repo.URI)
			// setRepoFieldsFromRemote is used in read requests, including
			// unauthed ones. However, this write isn't as the user, but
			// rather an optimization for us to save the data from
			// github. As such we use an elevated context to allow the
			// write.
			if err := localstore.Repos.Update(elevatedActor(ctx), *update); err != nil {
				log15.Error("Failed updating repo metadata from remote", "repo", repo.URI, "error", err)
			}
		}
	}
	if !repo.Mirror && repo.HTTPCloneURL == "" { // artifact of self-hosted repositories
		repo.HTTPCloneURL = conf.AppURL.ResolveReference(approuter.Rel.URLToRepo(repo.URI)).String()
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
func repoSetFromRemote(repo *sourcegraph.Repo, ghrepo *sourcegraph.Repo) *localstore.RepoUpdate {
	updated := false
	updateOp := &localstore.RepoUpdate{
		ReposUpdateOp: &sourcegraph.ReposUpdateOp{
			Repo: repo.ID,
		},
	}

	if ghrepo.URI != repo.URI {
		repo.URI = ghrepo.URI
		updateOp.URI = ghrepo.URI
		updated = true
	}
	if ghrepo.Owner != repo.Owner {
		repo.Owner = ghrepo.Owner
		updateOp.Owner = ghrepo.Owner
		updated = true
	}
	if ghrepo.Name != repo.Name {
		repo.Name = ghrepo.Name
		updateOp.Name = ghrepo.Name
		updated = true
	}
	if ghrepo.Description != repo.Description {
		repo.Description = ghrepo.Description
		updateOp.Description = ghrepo.Description
		updated = true
	}
	if ghrepo.HTTPCloneURL != repo.HTTPCloneURL {
		repo.HTTPCloneURL = ghrepo.HTTPCloneURL
		updateOp.HTTPCloneURL = ghrepo.HTTPCloneURL
		updated = true
	}
	if ghrepo.SSHCloneURL != repo.SSHCloneURL {
		repo.SSHCloneURL = ghrepo.SSHCloneURL
		updateOp.SSHCloneURL = ghrepo.SSHCloneURL
		updated = true
	}
	if ghrepo.HomepageURL != repo.HomepageURL {
		repo.HomepageURL = ghrepo.HomepageURL
		updateOp.HomepageURL = ghrepo.HomepageURL
		updated = true
	}
	if ghrepo.DefaultBranch != repo.DefaultBranch {
		repo.DefaultBranch = ghrepo.DefaultBranch
		updateOp.DefaultBranch = ghrepo.DefaultBranch
		updated = true
	}
	if ghrepo.Language != repo.Language {
		repo.Language = ghrepo.Language
		updateOp.Language = ghrepo.Language
		updated = true
	}
	if ghrepo.Blocked != repo.Blocked {
		repo.Blocked = ghrepo.Blocked
		if ghrepo.Blocked {
			updateOp.Blocked = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Blocked = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Deprecated != repo.Deprecated {
		repo.Deprecated = ghrepo.Deprecated
		if ghrepo.Deprecated {
			updateOp.Deprecated = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Deprecated = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Fork != repo.Fork {
		repo.Fork = ghrepo.Fork
		if ghrepo.Fork {
			updateOp.Fork = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Fork = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Mirror != repo.Mirror {
		repo.Mirror = ghrepo.Mirror
		if ghrepo.Mirror {
			updateOp.Mirror = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Mirror = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Private != repo.Private {
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

	if ghrepo.Origin != nil {
		if repo.Origin != nil {
			if repo.Origin.ID != ghrepo.Origin.ID || repo.Origin.Service != ghrepo.Origin.Service || repo.Origin.APIBaseURL != ghrepo.Origin.APIBaseURL {
				repo.Origin = ghrepo.Origin
				updateOp.Origin = ghrepo.Origin
				updated = true
			}
		} else {
			repo.Origin = ghrepo.Origin
			updateOp.Origin = ghrepo.Origin
			updated = true
		}
	} else {
		if repo.Origin != nil {
			repo.Origin = nil
			updateOp.Origin = &sourcegraph.Origin{}
			updated = true
		}
	}

	// The Permissions field should NOT be persisted, because it is
	// specific to the current user who requested the repo. So, don't
	// track updated.
	repo.Permissions = ghrepo.Permissions

	if updated {
		return updateOp
	}
	return nil
}

func (s *repos) Create(ctx context.Context, op *sourcegraph.ReposCreateOp) (*sourcegraph.Repo, error) {
	var repo *sourcegraph.Repo
	switch {
	case op.GetFromGitHubID() != 0:
		var err error
		repo, err = s.newRepoFromOrigin(ctx, &sourcegraph.Origin{
			Service: sourcegraph.Origin_GitHub,
			ID:      strconv.Itoa(int(op.GetFromGitHubID())),
		})
		if err != nil {
			return nil, err
		}

	case op.GetOrigin() != nil:
		var err error
		repo, err = s.newRepoFromOrigin(ctx, op.GetOrigin())
		if err != nil {
			return nil, err
		}

	// Intercept "github.com/..." repos, create them in the same manner as from GitHub repo ID.
	case op.GetNew() != nil && strings.HasPrefix(strings.ToLower(op.GetNew().URI), "github.com/"):
		op := op.GetNew()
		if !op.Mirror {
			return nil, grpc.Errorf(codes.InvalidArgument, "github.com/ repos can only be mirrors")
		}
		if !(op.CloneURL == fmt.Sprintf("https://%s", op.URI) || op.CloneURL == fmt.Sprintf("https://%s.git", op.URI)) {
			// Disallow creating GitHub mirrors via repo URI and clone URL unless they match.
			return nil, grpc.Errorf(codes.InvalidArgument, "github.com/ mirrors repos can only be created if the clone URL matches the repo URI")
		}

		// TODO: This is gross, but the current sourcegraph.Origin struct does not easily
		//       permit better (I've tried). It can be improved/simplified, but it's a more
		//       significant undertaking than what's in scope right now (so I'm purposefully
		//       choosing a less than ideal solution for now). It's more okay because it only
		//       happens when creating a GH repo via URI, which may get removed.
		//
		// Make a request to GH API simply to resolve the owner and repo into a GitHub repo ID,
		// discard the rest of the metadata, then use that as origin, which in turn makes
		// another call to GH API to fetch the same repo metadata.
		// The goal here is to avoid multiple distinct code paths for GH repo creation, because the
		// subtle business logic of how we process GH metadata changes, and it's easy for multiple
		// code paths to diverge.
		ghRepo, err := github.ReposFromContext(ctx).Get(ctx, op.URI)
		if err != nil {
			return nil, err
		}
		if ghRepo.Origin == nil {
			return nil, fmt.Errorf("unexpected nil Origin from our GitHub repo store")
		}
		if ghRepo.Origin.Service != sourcegraph.Origin_GitHub {
			return nil, fmt.Errorf("unexpected Origin.Service from our GitHub repo store: %v", ghRepo.Origin.Service)
		}

		repo, err = s.newRepoFromOrigin(ctx, &sourcegraph.Origin{
			Service: sourcegraph.Origin_GitHub,
			ID:      ghRepo.Origin.ID,
		})
		if err != nil {
			return nil, err
		}

	case op.GetNew() != nil:
		var err error
		repo, err = s.newRepo(ctx, op.GetNew())
		if err != nil {
			return nil, err
		}

	default:
		return nil, grpc.Errorf(codes.Unimplemented, "repo creation operation not supported")
	}

	repoID, err := localstore.Repos.Create(ctx, repo)
	if err != nil {
		return nil, err
	}
	repo.ID = repoID

	repo, err = s.Get(ctx, &sourcegraph.RepoSpec{ID: repo.ID})
	if err != nil {
		return nil, err
	}

	repoMaybeEnqueueUpdate(ctx, repo)
	sendCreateRepoSlackMsg(ctx, repo.URI, repo.Language, repo.Mirror, repo.Private)

	return repo, nil
}

// repoMaybeEnqueueUpdate enqueues an update as the current user if the repo
// is a Mirror.
func repoMaybeEnqueueUpdate(ctx context.Context, repo *sourcegraph.Repo) {
	if !repo.Mirror {
		return
	}
	var asUser *sourcegraph.UserSpec
	if actor := authpkg.ActorFromContext(ctx); actor.UID != "" {
		asUser = actor.UserSpec()
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
		return nil, grpc.Errorf(codes.InvalidArgument, "newRepo is not allowed to create github.com/ repos")
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
	update := localstore.RepoUpdate{ReposUpdateOp: op, UpdatedAt: &ts}
	if err := localstore.Repos.Update(ctx, update); err != nil {
		return nil, err
	}

	return s.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
}

func (s *repos) Delete(ctx context.Context, repo *sourcegraph.RepoSpec) (*pbtypes.Void, error) {
	if err := localstore.Repos.Delete(ctx, repo.ID); err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *repos) GetConfig(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
	conf, err := localstore.RepoConfigs.Get(ctx, repo.ID)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &sourcegraph.RepoConfig{}
	}
	return conf, nil
}

func (s *repos) GetInventory(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	// Cap GetInventory operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

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
	vcsrepo, err := localstore.RepoVCS.Open(ctx, repoRev.Repo)
	if err != nil {
		return nil, err
	}

	fs := vcs.FileSystem(vcsrepo, vcs.CommitID(repoRev.CommitID))
	inv, err := inventory.Scan(ctx, vfsutil.Walkable(ctxvfs.StripContext(fs), filepath.Join))
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

func (s *repos) EnableWebhook(ctx context.Context, op *sourcegraph.RepoWebhookOptions) (*pbtypes.Void, error) {
	if !github.HasAuthedUser(ctx) {
		return nil, errors.New("Unauthed user")
	}

	var err error
	ctx, err = github.NewContextWithAuthedClient(ctx)
	if err != nil {
		return nil, err
	}
	if err := github.ReposFromContext(ctx).CreateHook(ctx, op.URI, &gogithub.Hook{
		Name:   gogithub.String("web"),
		Events: []string{"push", "pull_request"},
		Config: map[string]interface{}{
			"url":          conf.AppURL.String() + "/.api/webhook/callback",
			"content_type": "json",
		},
		Active: gogithub.Bool(true),
	}); err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}
