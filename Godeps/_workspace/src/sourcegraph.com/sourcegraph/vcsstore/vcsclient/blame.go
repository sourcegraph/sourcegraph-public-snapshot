package vcsclient

import "sourcegraph.com/sourcegraph/go-vcs/vcs"

func (r *repository) BlameFile(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	url, err := r.url(RouteRepoBlameFile, map[string]string{"Path": path}, opt)
	if err != nil {
		return nil, err
	}

	req, err := r.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	var hunks []*vcs.Hunk
	if _, err := r.client.Do(req, &hunks); err != nil {
		return nil, err
	}

	return hunks, nil
}
