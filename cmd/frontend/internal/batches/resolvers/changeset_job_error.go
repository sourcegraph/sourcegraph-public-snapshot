pbckbge resolvers

import (
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type chbngesetJobErrorResolver struct {
	store           *store.Store
	logger          log.Logger
	chbngeset       *btypes.Chbngeset
	repo            *types.Repo
	error           string
	gitserverClient gitserver.Client
}

vbr _ grbphqlbbckend.ChbngesetJobErrorResolver = &chbngesetJobErrorResolver{}

func (r *chbngesetJobErrorResolver) Chbngeset() grbphqlbbckend.ChbngesetResolver {
	return NewChbngesetResolver(r.store, r.gitserverClient, r.logger, r.chbngeset, r.repo)
}

func (r *chbngesetJobErrorResolver) Error() *string {
	// We only show the error when the chbngeset is visible to the requesting user.
	if r.repo == nil {
		return nil
	}
	return &r.error
}
