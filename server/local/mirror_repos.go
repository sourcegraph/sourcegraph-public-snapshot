package local

import (
	"os/exec"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/AaronO/go-git-http"
	"github.com/prometheus/client_golang/prometheus"

	gogithub "github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

var MirrorRepos sourcegraph.MirrorReposServer = &mirrorRepos{}

type mirrorRepos struct{}

var _ sourcegraph.MirrorReposServer = (*mirrorRepos)(nil)

var activeGitGC = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "git",
	Name:      "active_gc",
	Help:      `Total number of "git gc" commands that are currently running.`,
})

func init() {
	prometheus.MustRegister(activeGitGC)
}

func (s *mirrorRepos) RefreshVCS(ctx context.Context, op *sourcegraph.MirrorReposRefreshVCSOp) (*pbtypes.Void, error) {
	r, err := store.ReposFromContext(ctx).Get(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): What if multiple goroutines or processes
	// simultaneously clone or update the same repo? Race conditions
	// probably, esp. on NFS.

	remoteOpts := vcs.RemoteOpts{}
	var gh *gogithub.Client
	// For private repos, supply auth from local auth store.
	if r.Private {
		token, err := s.getRepoAuthToken(elevatedActor(ctx), op.Repo.URI)
		if err != nil {
			return nil, grpc.Errorf(codes.Unavailable, "could not fetch credentials for %v: %v", op.Repo.URI, err)
		}
		remoteOpts.HTTPS = &vcs.HTTPSConfig{
			Pass: token,
		}
		gh = githubutil.Default.AuthedClient(token)
	} else {
		gh = githubutil.Default.UnauthedClient()
	}

	vcsRepo, err := store.RepoVCSFromContext(ctx).Open(ctx, r.URI)
	if err != nil {
		return nil, err
	}
	if err := s.updateRepo(ctx, r, vcsRepo, remoteOpts); err != nil {
		if err != vcs.ErrRepoNotExist {
			return nil, err
		}
		if err := s.cloneRepo(ctx, r, remoteOpts); err != nil {
			return nil, err
		}
	}

	// Update any of the repo's metadata that has changed on GitHub.
	ctx = github.NewContextWithClient(ctx, gh)
	ghRepos := github.Repos{}
	ghRepo, err := ghRepos.Get(ctx, r.URI)
	if err != nil {
		return nil, err
	}
	err = store.ReposFromContext(ctx).Update(ctx, &store.RepoUpdate{
		ReposUpdateOp: &sourcegraph.ReposUpdateOp{
			Repo: sourcegraph.RepoSpec{
				URI: op.Repo.URI,
			},
			DefaultBranch: ghRepo.DefaultBranch,
			Language:      ghRepo.Language,
			Description:   ghRepo.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

func (s *mirrorRepos) getRepoAuthToken(ctx context.Context, repo string) (string, error) {
	users, err := store.RepoPermsFromContext(ctx).ListRepoUsers(ctx, repo)
	if err != nil {
		return "", err
	}

	if users == nil || len(users) == 0 {
		return "", grpc.Errorf(codes.Unavailable, "repo has no user with access")
	}

	var token string
	for _, user := range users {
		extToken, err := svc.Auth(ctx).GetExternalToken(ctx, &sourcegraph.ExternalTokenRequest{UID: user})
		if err != nil {
			log15.Debug("No auth token found for user to fetch repo", "uid", users[0], "repo", repo, "error", err)
			continue
		}
		token = extToken.Token
	}

	if token == "" {
		return "", grpc.Errorf(codes.Unavailable, "no auth token found for repo")
	}
	return token, nil
}

func (s *mirrorRepos) cloneRepo(ctx context.Context, repo *sourcegraph.Repo, remoteOpts vcs.RemoteOpts) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "MirrorRepos.CloneRepo", repo.URI); err != nil {
		return err
	}

	err := store.RepoVCSFromContext(ctx).Clone(ctx, repo.URI, true, true, &store.CloneInfo{
		VCS:        repo.VCS,
		CloneURL:   repo.HTTPCloneURL,
		RemoteOpts: remoteOpts,
	})
	if err != nil {
		return err
	}

	// We've just cloned the repository, so kick off a build on the default
	// branch. This isn't needed for the fs backend because it initializes an
	// empty repository first and then proceeds to just updateRepo, thus skipping
	// this clone phase entirely.
	commit, err := svc.Repos(ctx).GetCommit(ctx, &sourcegraph.RepoRevSpec{
		RepoSpec: repo.RepoSpec(),
		Rev:      repo.DefaultBranch,
	})
	if err != nil {
		return err
	}
	_, err = svc.Builds(ctx).Create(ctx, &sourcegraph.BuildsCreateOp{
		Repo:     repo.RepoSpec(),
		CommitID: string(commit.ID),
		Branch:   repo.DefaultBranch,
		Config:   sourcegraph.BuildConfig{Queue: true},
	})
	if err != nil {
		log15.Warn("cloneRepo: failed to create build", "err", err, "repo", repo.URI, "commit", commit.ID, "branch", repo.DefaultBranch)
		return nil
	}
	log15.Debug("cloneRepo: build created", "repo", repo.URI, "branch", repo.DefaultBranch, "commit", commit.ID)
	return nil
}

func (s *mirrorRepos) updateRepo(ctx context.Context, repo *sourcegraph.Repo, vcsRepo vcs.Repository, remoteOpts vcs.RemoteOpts) error {
	// TODO: Need to detect new tags and copy git_transport.go in event publishing
	// behavior.

	// Grab the current revision of every branch.
	branches, err := vcsRepo.Branches(vcs.BranchesOptions{})
	if err != nil {
		return err
	}

	// Update everything.
	updateResult, err := vcsRepo.UpdateEverything(remoteOpts)
	if err != nil {
		return err
	}

	if len(updateResult.Changes) > 0 {
		go func() {
			activeGitGC.Inc()
			defer activeGitGC.Dec()
			gcCmd := exec.Command("git", "gc")
			gcCmd.Dir = vcsRepo.GitRootDir()
			gcCmd.Run() // ignore error
		}()
	}

	forcePushes := make(map[string]bool)
	for _, change := range updateResult.Changes {
		switch change.Op {
		case vcs.NewOp, vcs.ForceUpdatedOp:
			// Skip refs that aren't branches, such as GitHub
			// "refs/pull/123/head" and "refs/pull/123/merge" refs
			// that are created for each pull request. In the future
			// we may want to handle these, but skipping them for now
			// is good because otherwise when we add a new mirror
			// repo, builds and notifications are triggered for all
			// historical PRs.
			if repo.VCS == "git" && strings.HasPrefix(change.Branch, "refs/") {
				continue
			}

			// Determine the event type, and if it's a force push mark for later to
			// avoid additional work.
			eventType := events.GitCreateBranchEvent
			gitEventType := githttp.EventType(githttp.PUSH)
			if change.Op == vcs.ForceUpdatedOp {
				// Force push, remember for later.
				forcePushes[change.Branch] = true
				eventType = events.GitPushEvent
				gitEventType = githttp.PUSH_FORCE
			}

			// Determine the new branch head revision.
			head, err := vcsRepo.ResolveRevision(change.Branch)
			if err != nil {
				return err
			}

			// Publish the event.
			// TODO: what about GitPayload.ContentEncoding field?
			events.Publish(eventType, events.GitPayload{
				Actor:       authpkg.UserSpecFromContext(ctx),
				Repo:        repo.RepoSpec(),
				IgnoreBuild: change.Branch != repo.DefaultBranch,
				Event: githttp.Event{
					Type:   gitEventType,
					Commit: string(head),
					Branch: change.Branch,
					Last:   emptyGitCommitID,
					// TODO: specify Dir, Tag, Error and Request fields somehow?
				},
			})
		}
	}

	// Find all new commits on each branch.
	for _, oldBranch := range branches {
		if _, ok := forcePushes[oldBranch.Name]; ok {
			// Already handled above.
			continue
		}

		// Determine new branch head revision.
		head, err := vcsRepo.ResolveRevision(oldBranch.Name)
		if err == vcs.ErrRevisionNotFound {
			// Branch was deleted.
			// TODO: what about GitPayload.ContentEncoding field?
			events.Publish(events.GitDeleteBranchEvent, events.GitPayload{
				Actor: authpkg.UserSpecFromContext(ctx),
				Repo:  repo.RepoSpec(),
				Event: githttp.Event{
					Type:   githttp.PUSH,
					Commit: emptyGitCommitID,
					Branch: oldBranch.Name,
					// TODO: specify Dir, Tag, Error and Request fields somehow?
				},
			})
			continue
		} else if err != nil {
			return err
		}
		if head == oldBranch.Head {
			continue // No new commits.
		}

		// Publish an event for the new commits pushed.
		// TODO: what about GitPayload.ContentEncoding field?
		events.Publish(events.GitPushEvent, events.GitPayload{
			Actor: authpkg.UserSpecFromContext(ctx),
			Repo:  repo.RepoSpec(),
			Event: githttp.Event{
				Type:   githttp.PUSH,
				Commit: string(head),
				Last:   string(oldBranch.Head),
				Branch: oldBranch.Name,
				// TODO: specify Dir, Tag, Error and Request fields somehow?
			},
		})
	}
	return nil
}

func (s *mirrorRepos) GetUserData(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.UserMirrorData, error) {
	gd := &sourcegraph.UserMirrorData{
		URL:   githubcli.Config.URL(),
		Host:  githubcli.Config.Host() + "/",
		State: sourcegraph.UserMirrorsState_NotAllowed,
	}

	// Try adding user to waitlist.
	res, err := s.AddToWaitlist(ctx, &pbtypes.Void{})
	if err != nil {
		return nil, err
	}
	gd.State = res.State
	if gd.State != sourcegraph.UserMirrorsState_HasAccess && gd.State != sourcegraph.UserMirrorsState_OnWaitlist {
		return gd, nil
	}

	// Fetch the currently authenticated user's stored access token (if any).
	extToken, err := svc.Auth(ctx).GetExternalToken(ctx, nil)
	if grpc.Code(err) == codes.NotFound {
		gd.State = sourcegraph.UserMirrorsState_NoToken
		return gd, nil
	} else if err != nil {
		return nil, err
	}

	repoPermsStore := store.RepoPermsFromContext(ctx)
	reposStore := store.ReposFromContext(ctx)

	ghRepos := &github.Repos{}
	// TODO(perf) Cache this response or perform the fetch after page load to avoid
	// having to wait for an http round trip to github.com.
	gitHubRepos, err := ghRepos.ListWithToken(ctx, extToken.Token)
	if err != nil {
		// Since the error is caused by something other than the token not existing,
		// ensure the user knows there is a value set for the token but that
		// it is invalid.
		gd.State = sourcegraph.UserMirrorsState_InvalidToken
		return gd, nil
	}

	var gitHubRepoURIs []string
	for _, repo := range gitHubRepos {
		if repo != nil {
			gitHubRepoURIs = append(gitHubRepoURIs, repo.URI)
		}
	}

	// This map contains repos visible to this GitHub user that are already
	// mirrored on this Sourcegraph server.
	existingRepos := make(map[string]*sourcegraph.Repo)
	repoOpts := &sourcegraph.RepoListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: 1000,
			Page:    1,
		},
		URIs: gitHubRepoURIs,
	}
	for {
		repoList, err := reposStore.List(elevatedActor(ctx), repoOpts)
		if err != nil {
			return nil, err
		}
		if len(repoList) == 0 {
			break
		}

		for _, repo := range repoList {
			existingRepos[repo.URI] = repo
		}

		repoOpts.ListOptions.Page += 1
	}

	// Check if a user's remote GitHub repo already exists locally under the
	// same URI. Allow this user to access all their private repos that are
	// already mirrored on this Sourcegraph.
	reposByOrg := make(map[string]*sourcegraph.RemoteOrgRepos)
	defaultOrgName := "public-org"
	var privateRepoURIs []string
	for _, repo := range gitHubRepos {
		if localRepo, ok := existingRepos[repo.URI]; ok {
			repo.ExistsLocally = true
			if localRepo.Private != repo.Private {
				// This means that either the existing local mirror was created with
				// the incorrect visibility (public vs. private) compared to the
				// upstream GitHub repo, or the upstream repo's visibility was modified
				// recently.
				// Modify the repo metadata in the local db to be consistent with it's
				// visibility on GitHub. This means that if the repo was public and is
				// now private, all Sourcegraph users will need to visit their dashboard
				// to force a sync with GitHub and have their repo_perms records updated
				// in order to continue accessing this repo.
				now := time.Now()
				err := reposStore.Update(elevatedActor(ctx), &store.RepoUpdate{
					ReposUpdateOp: &sourcegraph.ReposUpdateOp{
						IsPrivate: repo.Private,
						IsPublic:  !repo.Private,
					},
					UpdatedAt: &now,
				})
				if err != nil {
					// Failed to update repo visibility, so skip it.
					log15.Warn("Failed to update repo visibility level", "repo", repo.URI, "error", err)
					continue
				}

				// If the repo is public, remove it's records from the repo_perms store
				// which only stores permissions for private repos.
				if !repo.Private {
					err := repoPermsStore.DeleteRepo(elevatedActor(ctx), repo.URI)
					if err != nil {
						// Failed to update repo permissions, so skip it.
						log15.Warn("Failed to update repo permissions", "repo", repo.URI, "error", err)
						continue
					}
				}
			}
		} else if repo.Private {
			// If repo is private and not mirrored yet, add it to pending repos.
			if err := store.WaitlistFromContext(ctx).RecordPendingRepo(elevatedActor(ctx), repo); err != nil {
				log15.Warn("Failed to record pending private repo", "uri", repo.URI, "error", err)
			}
		}
		orgName := defaultOrgName
		if repo.Owner != nil {
			orgName = repo.Owner.Login
		}
		if _, ok := reposByOrg[orgName]; !ok {
			reposByOrg[orgName] = &sourcegraph.RemoteOrgRepos{
				PrivateRepos: make([]*sourcegraph.RemoteRepo, 0),
				PublicRepos:  make([]*sourcegraph.RemoteRepo, 0),
			}
		}

		if repo.Private {
			reposByOrg[orgName].PrivateRepos = append(reposByOrg[orgName].PrivateRepos, repo)
			privateRepoURIs = append(privateRepoURIs, repo.URI)
		} else {
			reposByOrg[orgName].PublicRepos = append(reposByOrg[orgName].PublicRepos, repo)
		}
	}
	// Update the user's permissions for visible private repos.
	uid := int32(authpkg.ActorFromContext(ctx).UID)
	if err := repoPermsStore.Update(elevatedActor(ctx), uid, privateRepoURIs); err != nil {
		log15.Error("Failed to set private repo permissions for user", "uid", uid, "error", err)
	}
	gd.ReposByOrg = reposByOrg

	return gd, nil
}

func (s *mirrorRepos) AddToWaitlist(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.WaitlistState, error) {
	uid := int32(authpkg.ActorFromContext(ctx).UID)
	userLogin := authpkg.ActorFromContext(ctx).Login

	if uid == 0 {
		return nil, grpc.Errorf(codes.Unauthenticated, "no authenticated user found in context")
	}

	result := &sourcegraph.WaitlistState{State: sourcegraph.UserMirrorsState_NotAllowed}

	if authutil.ActiveFlags.MirrorsWaitlist == "none" {
		result.State = sourcegraph.UserMirrorsState_HasAccess
		return result, nil
	}

	// Fetch the currently authenticated user's stored access token (if any).
	extToken, err := svc.Auth(ctx).GetExternalToken(ctx, &sourcegraph.ExternalTokenRequest{})
	if grpc.Code(err) == codes.NotFound {
		result.State = sourcegraph.UserMirrorsState_NoToken
		return result, nil
	} else if err != nil {
		return nil, err
	}

	client := githubutil.Default.AuthedClient(extToken.Token)
	ghUser, _, err := client.Users.Get("")
	if err != nil {
		result.State = sourcegraph.UserMirrorsState_InvalidToken
		return result, nil
	}

	waitlistStore := store.WaitlistFromContext(ctx)

	waitlistedUser, err := waitlistStore.GetUser(ctx, uid)
	if err != nil {
		if _, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
			return nil, err
		}
		// User is not waitlisted
		waitlistedUser = &sourcegraph.WaitlistedUser{UID: uid}
	}

	if waitlistedUser.GrantedAt != nil {
		// User has been granted access already
		result.State = sourcegraph.UserMirrorsState_HasAccess
		return result, nil
	}

	if waitlistedUser.AddedAt == nil {
		// User is not on the waitlist, so add them.
		if err := waitlistStore.AddUser(ctx, uid); err != nil {
			return nil, err
		}
	}
	result.State = sourcegraph.UserMirrorsState_OnWaitlist

	// User is on the waitlist, but check if any of their orgs has been granted access already.
	ghOrgs, _, err := client.Organizations.List("", nil)
	if err != nil {
		log15.Error("Could not list GitHub orgs for user", "github_user", *ghUser.Login, "sourcegraph_user", userLogin, "error", err)
		return result, nil
	}

	orgNames := make([]string, len(ghOrgs))
	for i, org := range ghOrgs {
		orgNames[i] = *org.Login

		// Add org to waitlist if it doesn't exist already.
		err := waitlistStore.AddOrg(elevatedActor(ctx), orgNames[i])
		if err != nil && err != store.ErrWaitlistedOrgExists {
			return nil, err
		}
	}
	grantedOrgs, err := waitlistStore.ListOrgs(elevatedActor(ctx), false, true, orgNames)
	if err != nil {
		log15.Error("Could not check waitlisted orgs for user", "github_user", *ghUser.Login, "sourcegraph_user", userLogin, "error", err)
		return result, nil
	}
	if len(grantedOrgs) > 0 {
		// User should be granted access automatically.
		if err := waitlistStore.GrantUser(elevatedActor(ctx), uid); err != nil {
			return nil, err
		}
		result.State = sourcegraph.UserMirrorsState_HasAccess
	}

	return result, nil
}
