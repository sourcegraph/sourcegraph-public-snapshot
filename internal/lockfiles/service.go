package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	GitArchive func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error)
}

func (s *Service) StreamDependencies(ctx context.Context, repo api.RepoName, rev string, cb func(*Dependency) error) (err error) {
	tr, ctx := trace.New(ctx, "lockfiles.StreamDependencies", string(repo)+"@"+rev)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	opts := gitserver.ArchiveOptions{
		Treeish: rev,
		Format:  "zip",
		Paths: []string{
			"*" + NPMFilename,
		},
	}

	rc, err := s.GitArchive(ctx, repo, opts)
	if err != nil {
		return err
	}

	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
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
			k := d.String()
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

func (s *Service) ListDependencies(ctx context.Context, repo api.RepoName, rev string) (deps []*Dependency, err error) {
	tr, ctx := trace.New(ctx, "lockfiles.ListDependencies", string(repo)+"@"+rev)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	err = s.StreamDependencies(ctx, repo, rev, func(d *Dependency) error {
		deps = append(deps, d)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.SliceStable(deps, func(i, j int) bool { return deps[i].Less(deps[j]) })
	return deps, nil
}

func parseZipLockfile(f *zip.File) ([]*Dependency, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ds, err := Parse(f.Name, contents)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
