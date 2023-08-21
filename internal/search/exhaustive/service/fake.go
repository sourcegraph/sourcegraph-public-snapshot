package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewSearcherFake is a convenient working implementation of SearchQuery which
// always will write results generated from the repoRevs. It expects a query
// string which looks like
//
//	 1@rev1 1@rev2 2@rev3
//
//	This is a space separated list of {repoid}@{revision}.
//
//	- RepositoryRevSpecs will return one RepositoryRevSpec per unique repository.
//	- ResolveRepositoryRevSpec returns the repoRevs for that repository.
//	- Search will write one result which is just the repo and revision.
func NewSearcherFake() NewSearcher {
	return backendFake{}
}

type backendFake struct{}

func (backendFake) NewSearch(ctx context.Context, q string) (SearchQuery, error) {
	var repoRevs []RepositoryRevision
	for _, part := range strings.Fields(q) {
		var r RepositoryRevision
		if n, err := fmt.Sscanf(part, "%d@%s", &r.Repository, &r.Revision); n != 2 || err != nil {
			return nil, errors.Errorf("failed to parse repository revision %q", part)
		}
		r.RepositoryRevSpec.Repository = r.Repository
		r.RepositoryRevSpec.RevisionSpecifier = "spec"
		repoRevs = append(repoRevs, r)
	}
	return searcherFake{repoRevs: repoRevs}, nil
}

type searcherFake struct {
	repoRevs []RepositoryRevision
}

func (s searcherFake) RepositoryRevSpecs(context.Context) ([]RepositoryRevSpec, error) {
	seen := map[RepositoryRevSpec]bool{}
	var repoRevSpecs []RepositoryRevSpec
	for _, r := range s.repoRevs {
		if seen[r.RepositoryRevSpec] {
			continue
		}
		seen[r.RepositoryRevSpec] = true
		repoRevSpecs = append(repoRevSpecs, r.RepositoryRevSpec)
	}
	return repoRevSpecs, nil
}

func (s searcherFake) ResolveRepositoryRevSpec(_ context.Context, repoRevSpec RepositoryRevSpec) ([]RepositoryRevision, error) {
	var repoRevs []RepositoryRevision
	for _, r := range s.repoRevs {
		if r.RepositoryRevSpec == repoRevSpec {
			repoRevs = append(repoRevs, r)
		}
	}
	return repoRevs, nil
}

func (s searcherFake) Search(_ context.Context, r RepositoryRevision, w CSVWriter) error {
	if err := w.WriteHeader("repo", "revspec", "revision"); err != nil {
		return err
	}
	return w.WriteRow(strconv.Itoa(int(r.Repository)), r.RevisionSpecifier, string(r.Revision))
}
