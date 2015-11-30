package local

import (
	"os"

	"github.com/AaronO/go-git-http"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var MirrorRepos sourcegraph.MirrorReposServer = &mirrorRepos{}

type mirrorRepos struct{}

var _ sourcegraph.MirrorReposServer = (*mirrorRepos)(nil)

func (s *mirrorRepos) RefreshVCS(ctx context.Context, op *sourcegraph.MirrorReposRefreshVCSOp) (*pbtypes.Void, error) {
	r, err := store.ReposFromContext(ctx).Get(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): What if multiple goroutines or processes
	// simultaneously clone or update the same repo? Race conditions
	// probably, esp. on NFS.

	remoteOpts := vcs.RemoteOpts{}
	if op.Credentials != nil {
		remoteOpts.HTTPS = &vcs.HTTPSConfig{
			Pass: op.Credentials.Pass,
		}
	}

	vcsRepo, err := store.RepoVCSFromContext(ctx).Open(ctx, r.URI)
	if os.IsNotExist(err) || grpc.Code(err) == codes.NotFound {
		err = s.cloneRepo(ctx, r, remoteOpts)
	} else if err != nil {
		return nil, err
	} else {
		err = s.updateRepo(ctx, r, vcsRepo, remoteOpts)
	}
	if err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

func (s *mirrorRepos) cloneRepo(ctx context.Context, repo *sourcegraph.Repo, remoteOpts vcs.RemoteOpts) error {
	return store.RepoVCSFromContext(ctx).Clone(ctx, repo.URI, &vcsclient.CloneInfo{
		VCS:        repo.VCS,
		CloneURL:   repo.HTTPCloneURL,
		RemoteOpts: remoteOpts,
	})
}

func (s *mirrorRepos) updateRepo(ctx context.Context, repo *sourcegraph.Repo, vcsRepo vcs.Repository, remoteOpts vcs.RemoteOpts) error {
	ru, ok := vcsRepo.(vcs.RemoteUpdater)
	if !ok {
		return &sourcegraph.NotImplementedError{What: "MirrorRepos.RefreshVCS on hosted repo"}
	}

	// TODO: Need to detect new branches and copy git_transport.go in event
	// publishing behavior.
	//
	// TODO: Need to detect new tags and copy git_transport.go in event publishing
	// behavior.

	// Get the current revision of every branch so we can build a list of commits
	// that have been pushed to each branch during the update.
	branches, err := vcsRepo.Branches(vcs.BranchesOptions{})
	if err != nil {
		return err
	}

	// Update everything.
	if err := ru.UpdateEverything(remoteOpts); err != nil {
		return err
	}

	// Find all new commits on each branch.
	for _, branch := range branches {
		// Determine new branch head revision.
		head, err := vcsRepo.ResolveBranch(branch.Name)
		if err != nil {
			return err
		}

		// Grab just new commits.
		commits, _, err := vcsRepo.Commits(vcs.CommitsOptions{
			Base:    branch.Head,
			Head:    head,
			NoTotal: true,
		})
		if err != nil {
			return err
		}

		// Publish an event for each commit pushed.
		for _, commit := range commits {
			// Resolve the last commit behind
			lastCommit, _, err := vcsRepo.Commits(vcs.CommitsOptions{
				Head:    commit.ID,
				N:       1,
				NoTotal: true,
			})
			if err != nil {
				return err
			}

			// TODO: what about GitPayload.ContentEncoding field?
			events.Publish(events.GitPushEvent, events.GitPayload{
				Actor: authpkg.UserSpecFromContext(ctx),
				Repo:  repo.RepoSpec(),
				Event: githttp.Event{
					Type:   githttp.PUSH, // TODO: detect githttp.PUSH_FORCE somehow?
					Commit: string(commit.ID),
					Last:   string(lastCommit[0].ID),
					Branch: branch.Name,
					// TODO: specify Dir, Tag, Error and Request fields somehow?
				},
			})
		}
	}
	return nil
}
