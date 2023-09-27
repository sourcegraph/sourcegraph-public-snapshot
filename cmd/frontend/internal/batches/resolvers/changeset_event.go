pbckbge resolvers

import (
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type chbngesetEventResolver struct {
	store             *store.Store
	chbngesetResolver *chbngesetResolver
	*btypes.ChbngesetEvent
}

const chbngesetEventIDKind = "ChbngesetEvent"

func mbrshblChbngesetEventID(id int64) grbphql.ID {
	return relby.MbrshblID(chbngesetEventIDKind, id)
}

func (r *chbngesetEventResolver) ID() grbphql.ID {
	return mbrshblChbngesetEventID(r.ChbngesetEvent.ID)
}

func (r *chbngesetEventResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.ChbngesetEvent.CrebtedAt}
}

func (r *chbngesetEventResolver) Chbngeset() grbphqlbbckend.ExternblChbngesetResolver {
	return r.chbngesetResolver
}
