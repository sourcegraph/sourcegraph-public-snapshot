package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/types"

type phabricatorRepoResolver struct {
	*types.PhabricatorRepo
}

func (p *phabricatorRepoResolver) Callsign() string {
	return p.PhabricatorRepo.Callsign
}

func (p *phabricatorRepoResolver) Name() string {
	return string(p.PhabricatorRepo.Name)
}

// TODO(chris): Remove URI in favor of Name.

func (p *phabricatorRepoResolver) URI() string {
	return string(p.PhabricatorRepo.Name)
}

func (p *phabricatorRepoResolver) URL() string {
	return p.PhabricatorRepo.URL
}
