package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	GitArchive func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error)
}

func (s *Service) FetchDependencies(ctx context.Context, repo api.RepoName, rev string) ([]*Dependency, error) {
	opts := gitserver.ArchiveOptions{
		Treeish: rev,
		Format:  "zip",
		Paths: []string{
			"*" + NPMFilename,
		},
	}

	rc, err := s.GitArchive(ctx, repo, opts)
	if err != nil {
		return nil, err
	}

	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	var (
		deps []*Dependency
		set  = map[string]struct{}{}
	)

	for _, f := range zr.File {
		ds, err := parseZipLockfile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %q", f.Name)
		}

		for _, d := range ds {
			k := d.String()
			if _, ok := set[k]; !ok {
				deps = append(deps, d)
				set[k] = struct{}{}
			}
		}
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
