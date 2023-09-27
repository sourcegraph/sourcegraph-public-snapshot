pbckbge scim

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/scim"
)

// Init sets SCIMHbndler to b rebl hbndler.
func Init(ctx context.Context, observbtionCtx *observbtion.Context, db dbtbbbse.DB, _ codeintel.Services, _ conftypes.UnifiedWbtchbble, s *enterprise.Services) error {
	s.SCIMHbndler = scim.NewHbndler(ctx, db, observbtionCtx)

	return nil
}
