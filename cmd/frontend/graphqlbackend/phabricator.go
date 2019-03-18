package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type phabricatorRepoResolver struct {
	*types.PhabricatorRepo
}

func (p *phabricatorRepoResolver) Callsign() string {
	return p.PhabricatorRepo.Callsign
}

func (p *phabricatorRepoResolver) Name() string {
	return string(p.PhabricatorRepo.URI)
}

// TODO(chris): Remove URI in favor of Name.
func (p *phabricatorRepoResolver) URI() string {
	return string(p.PhabricatorRepo.URI)
}

func (p *phabricatorRepoResolver) URL() string {
	return p.PhabricatorRepo.URL
}
