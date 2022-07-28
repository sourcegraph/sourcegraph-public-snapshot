package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"path"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	gitSvc     GitService
	operations *operations
}

func newService(gitSvc GitService, observationContext *observation.Context) *Service {
	return &Service{
		gitSvc:     gitSvc,
		operations: newOperations(observationContext),
	}
}

type Result struct {
	// Lockfile is the name of the lockfile that was parsed to get the list of
	// Deps and, optionally, the DependencyGraph.
	Lockfile string

	// Deps is the flat list of all dependencies found in Lockfile.
	Deps []reposource.VersionedPackage
	// Graph is the dependency graph found in the Lockfile. If no graph could
	// be built (`package.json` without `yarn.lock` doesn't allow building a
	// fully-resolved dependency graph).
	Graph *DependencyGraph
}

func (s *Service) ListDependencies(ctx context.Context, repo api.RepoName, rev string) (results []Result, err error) {
	ctx, _, endObservation := s.operations.listDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(repo)),
		log.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	err = s.StreamDependencies(ctx, repo, rev, func(r Result) error {
		results = append(results, r)
		return nil
	})

	return results, err
}

func (s *Service) StreamDependencies(ctx context.Context, repo api.RepoName, rev string, cb func(Result) error) (err error) {
	ctx, _, endObservation := s.operations.streamDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(repo)),
		log.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	// First call ls-files to find matching lockfiles, then pass those literal paths to archive.
	//
	// This ls-files call might appear redundant with the subsequent archive call, but it turns out
	// ls-files handles pathspecs that match 0 files whereas archive throws an error. The ls-files
	// behavior is desirable because we don't know ahead of time which lockfiles a repo has.
	paths, err := s.gitSvc.LsFiles(ctx, repo, api.CommitID(rev), lockfilePathspecs...)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		return nil
	}

	pathspecs := []gitdomain.Pathspec{}
	for _, p := range paths {
		pathspecs = append(pathspecs, gitdomain.PathspecLiteral(p))
	}

	opts := gitserver.ArchiveOptions{
		Treeish:   rev,
		Format:    gitserver.ArchiveFormatZip,
		Pathspecs: pathspecs,
	}

	rc, err := s.gitSvc.Archive(ctx, repo, opts)
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		if strings.Contains(err.Error(), "did not match any files") {
			return nil
		}
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if f.Mode().IsDir() {
			continue
		}

		ds, graph, err := parseZipLockfile(f)
		if err != nil {
			return errors.Wrapf(err, "failed to parse %q", f.Name)
		}

		result := Result{Lockfile: f.Name, Graph: graph}

		set := make(map[string]struct{})
		for _, d := range ds {
			k := d.VersionedPackageSyntax()
			if _, ok := set[k]; !ok {
				set[k] = struct{}{}
				result.Deps = append(result.Deps, d)
			}
		}

		cb(result)
	}

	return nil
}

func parseZipLockfile(f *zip.File) ([]reposource.VersionedPackage, *DependencyGraph, error) {
	r, err := f.Open()
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	logger := sglog.Scoped("parseZipLockfile", "")

	deps, graph, err := parse(f.Name, r)
	if err != nil {
		logger.Warn("failed to parse some lockfile dependencies", sglog.Error(err), sglog.String("file", f.Name))
	}

	return deps, graph, nil
}

func parse(file string, r io.Reader) ([]reposource.VersionedPackage, *DependencyGraph, error) {
	parser, ok := parsers[path.Base(file)]
	if !ok {
		return nil, nil, ErrUnsupported
	}
	return parser(r)
}
