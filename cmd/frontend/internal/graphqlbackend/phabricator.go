package graphqlbackend

import (
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type phabricatorRepoResolver struct {
	*sourcegraph.PhabricatorRepo
}

func (p *phabricatorRepoResolver) Callsign() string {
	return p.PhabricatorRepo.Callsign
}

func (p *phabricatorRepoResolver) URI() string {
	return p.PhabricatorRepo.URI
}
