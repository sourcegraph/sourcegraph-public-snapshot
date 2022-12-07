package intel

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	stores "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IntelService struct {
	bstore *store.Store
	ciserv codeintel.Services
	client graphql.Client
	logger log.Logger
}

func NewIntelService(logger log.Logger, bstore *store.Store, octx *observation.Context) (*IntelService, error) {
	cs, err := codeintel.NewServices(codeintel.ServiceDependencies{
		DB: bstore.DatabaseDB(),
		// TODO: This doesn't work when the codeintel db is a separate DB.
		CodeIntelDB:    stores.NewCodeIntelDBWith(bstore.DatabaseDB()),
		ObservationCtx: octx,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating code intel services")
	}

	return &IntelService{
		bstore: bstore,
		ciserv: cs,
		client: graphql.NewClient(internalapi.Client.URL+"/.internal/graphql", httpcli.InternalClient),
		logger: logger,
	}, nil
}
