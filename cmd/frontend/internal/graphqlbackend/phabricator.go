package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

type phabricatorRepoResolver struct {
	*types.PhabricatorRepo
}

func (p *phabricatorRepoResolver) Callsign() string {
	return p.PhabricatorRepo.Callsign
}

func (p *phabricatorRepoResolver) URI() string {
	return p.PhabricatorRepo.URI
}

func (p *phabricatorRepoResolver) URL() string {
	return p.PhabricatorRepo.URL
}
