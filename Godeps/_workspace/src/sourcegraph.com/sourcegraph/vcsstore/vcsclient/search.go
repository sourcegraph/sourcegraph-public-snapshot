package vcsclient

import "sourcegraph.com/sourcegraph/go-vcs/vcs"

func (r *repository) Search(at vcs.CommitID, opt vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	url, err := r.url(RouteRepoSearch, map[string]string{"CommitID": string(at)}, opt)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var res []*vcs.SearchResult
	if _, err := r.client.Do(req, &res); err != nil {
		return nil, err
	}

	return res, nil
}
