package dependencies

import (
	"context"
	"io"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var (
	svc     *Service
	svcOnce sync.Once
)

func GetService(db database.DB, syncer Syncer) *Service {
	svcOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		svc = newService(
			store.GetStore(db),
			lockfiles.GetService(&gitService{authz.DefaultSubRepoPermsChecker}),
			syncer,
			observationContext,
		)
	})

	return svc
}

type gitService struct {
	checker authz.SubRepoPermissionChecker
}

func (s *gitService) LsFiles(ctx context.Context, repo api.RepoName, commits api.CommitID, paths ...string) ([]string, error) {
	return git.LsFiles(ctx, s.checker, repo, commits, paths...)
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	return gitserver.DefaultClient.Archive(ctx, repo, opts)
}
