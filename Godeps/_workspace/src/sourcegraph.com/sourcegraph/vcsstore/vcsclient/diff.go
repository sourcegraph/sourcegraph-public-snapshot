package vcsclient

import (
	"fmt"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

var (
	_ vcs.Differ          = (*repository)(nil)
	_ vcs.CrossRepoDiffer = (*repository)(nil)
)

func (r *repository) Diff(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	url, err := r.url(RouteRepoDiff, map[string]string{"Base": string(base), "Head": string(head)}, opt)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var diff *vcs.Diff
	if _, err := r.client.Do(req, &diff); err != nil {
		return nil, err
	}

	return diff, nil
}

func (r *repository) CrossRepoDiff(base vcs.CommitID, headRepo vcs.Repository, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	// Only support cross-repo diffing for repos that we know how to
	// introspect.
	headRepo2, ok := headRepo.(*repository)
	if !ok {
		return nil, fmt.Errorf("cross-repo diffing in vcsclient is not implemented for %T", headRepo)
	}

	url, err := r.url(RouteRepoCrossRepoDiff, map[string]string{"Base": string(base), "HeadRepoPath": headRepo2.repoPath, "Head": string(head)}, opt)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var diff *vcs.Diff
	if _, err := r.client.Do(req, &diff); err != nil {
		return nil, err
	}

	return diff, nil
}
