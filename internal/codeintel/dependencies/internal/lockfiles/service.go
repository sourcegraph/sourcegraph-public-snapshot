package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"path"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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

func (s *Service) ListDependencies(ctx context.Context, repo api.RepoName, rev string) (deps []reposource.PackageDependency, err error) {
	ctx, _, endObservation := s.operations.listDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(repo)),
		log.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	err = s.StreamDependencies(ctx, repo, rev, func(d reposource.PackageDependency) error {
		deps = append(deps, d)
		return nil
	})

	return deps, err
}

func (s *Service) StreamDependencies(ctx context.Context, repo api.RepoName, rev string, cb func(reposource.PackageDependency) error) (err error) {
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

	pathspecs := []gitserver.Pathspec{}
	for _, p := range paths {
		pathspecs = append(pathspecs, gitserver.PathspecLiteral(p))
	}

	opts := gitserver.ArchiveOptions{
		Treeish:   rev,
		Format:    "zip",
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

	set := map[string]struct{}{}
	for _, f := range zr.File {
		if f.Mode().IsDir() {
			continue
		}

		ds, err := parseZipLockfile(f)
		if err != nil {
			return errors.Wrapf(err, "failed to parse %q", f.Name)
		}

		for _, d := range ds {
			k := d.PackageManagerSyntax()
			if _, ok := set[k]; !ok {
				set[k] = struct{}{}
				if err = cb(d); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func parseZipLockfile(f *zip.File) ([]reposource.PackageDependency, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	deps, err := parse(f.Name, r)
	if err != nil {
		log15.Warn("failed to parse some lockfile dependencies", "error", err, "file", f.Name)
	}

	return deps, nil
}

func parse(file string, r io.Reader) ([]reposource.PackageDependency, error) {
	parser, ok := parsers[path.Base(file)]
	if !ok {
		return nil, ErrUnsupported
	}
	return parser(r)
}
