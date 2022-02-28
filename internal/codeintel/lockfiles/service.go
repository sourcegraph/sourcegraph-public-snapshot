package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/authz"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	checker    authz.SubRepoPermissionChecker
	lsFiles    func(context.Context, authz.SubRepoPermissionChecker, api.RepoName, api.CommitID, ...string) ([]string, error)
	archive    func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error)
	operations *operations
}

func NewService(
	checker authz.SubRepoPermissionChecker,
	lsFiles func(context.Context, authz.SubRepoPermissionChecker, api.RepoName, api.CommitID, ...string) ([]string, error),
	archive func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error),
	observationContext *observation.Context,
) *Service {
	return &Service{
		checker:    checker,
		lsFiles:    lsFiles,
		archive:    archive,
		operations: newOperations(observationContext),
	}
}

func (s *Service) StreamDependencies(ctx context.Context, repo api.RepoName, rev string, cb func(reposource.PackageDependency) error) (err error) {
	ctx, endObservation := s.operations.streamDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(repo)),
		log.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	paths, err := s.lsFiles(ctx, s.checker, repo, api.CommitID(rev), lockfilePaths...)
	if err != nil {
		return err
	}

	opts := gitserver.ArchiveOptions{
		Treeish: rev,
		Format:  "zip",
		Paths:   paths,
	}

	rc, err := s.archive(ctx, repo, opts)
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

func (s *Service) ListDependencies(ctx context.Context, repo api.RepoName, rev string) (deps []reposource.PackageDependency, err error) {
	ctx, endObservation := s.operations.listDependencies.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

var lockfilePaths = func() (paths []string) {
	paths = make([]string, 0, len(parsers))
	for filename := range parsers {
		paths = append(paths, "*"+filename)
	}
	sort.Strings(paths)
	return
}()

func parseZipLockfile(f *zip.File) (deps []reposource.PackageDependency, err error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	ds, err := Parse(f.Name, r)
	if err != nil {
		log15.Warn("failed to parse some lockfile dependencies", "error", err, "file", f.Name)
	}

	return ds, nil
}
