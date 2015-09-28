package vcsclient

import (
	"fmt"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

var (
	_ vcs.Merger          = (*repository)(nil)
	_ vcs.CrossRepoMerger = (*repository)(nil)
)

func (r *repository) MergeBase(a, b vcs.CommitID) (vcs.CommitID, error) {
	url, err := r.url(RouteRepoMergeBase, map[string]string{"CommitIDA": string(a), "CommitIDB": string(b)}, nil)
	if err != nil {
		return "", err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.doIgnoringRedirects(req)
	if err != nil {
		return "", err
	}

	return r.parseCommitIDInURL(resp.Header.Get("location"))
}

func (r *repository) CrossRepoMergeBase(a vcs.CommitID, repoB vcs.Repository, b vcs.CommitID) (vcs.CommitID, error) {
	// Only support cross-repo ops for repos that we know how to
	// introspect.
	repoB2, ok := repoB.(*repository)
	if !ok {
		return "", fmt.Errorf("cross-repo merge-base in vcsclient is not implemented for %T", repoB)
	}

	url, err := r.url(RouteRepoCrossRepoMergeBase, map[string]string{"CommitIDA": string(a), "BRepoPath": repoB2.repoPath, "CommitIDB": string(b)}, nil)
	if err != nil {
		return "", err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.doIgnoringRedirects(req)
	if err != nil {
		return "", err
	}

	return r.parseCommitIDInURL(resp.Header.Get("location"))
}
