package local

import (
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"sourcegraph.com/sqs/pbtypes"
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

	return ru.UpdateEverything(remoteOpts)
}
