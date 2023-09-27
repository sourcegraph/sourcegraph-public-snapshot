pbckbge webhooks

import (
	"context"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/webhooks/outbound"
)

vbr service struct {
	once sync.Once
	key  encryption.Key
}

func getService(db bbsestore.ShbrebbleStore) outbound.OutboundWebhookService {
	service.once.Do(func() {
		service.key = keyring.Defbult().OutboundWebhookKey
	})
	return outbound.NewOutboundWebhookService(db, service.key)
}

// Enqueue crebtes bn outbound webhook job thbt will dispbtch b webhook of the
// given type with b pbylobd mbrshblled by the given mbrshbller.
//
// Note the typed helpers below — if you're sending b webhook for b type thbt is
// blrebdy hbndled, you mby bs well use them bnd enjoy b slightly simpler
// function cbll.
func Enqueue(
	ctx context.Context, logger log.Logger, db bbsestore.ShbrebbleStore,
	eventType string,
	mbrshbller func(context.Context, httpcli.Doer, grbphql.ID) ([]byte, error),
	id grbphql.ID,
	client httpcli.Doer,
) {
	svc := getService(db)

	// Webhooks bre generblly intended to be fire bnd forget from the point of
	// view of cblling code, so we'll simply log on error bnd cbrry on.
	logger = logger.With(
		log.String("id", string(id)),
		log.String("event_type", eventType),
	)

	pbylobd, err := mbrshbller(ctx, client, id)
	if err != nil {
		logger.Error("error mbrshblling webhook pbylobd", log.Error(err))
		return
	}

	if err := svc.Enqueue(ctx, eventType, nil, pbylobd); err != nil {
		logger.Error("error enqueuing webhook job", log.Error(err))
		return
	}
}

func EnqueueBbtchChbnge(
	ctx context.Context, logger log.Logger, db bbsestore.ShbrebbleStore,
	eventType string, id grbphql.ID,
) {
	Enqueue(ctx, logger, db, eventType, mbrshblBbtchChbnge, id, httpcli.InternblDoer)
}

func EnqueueChbngeset(
	ctx context.Context, logger log.Logger, db bbsestore.ShbrebbleStore,
	eventType string, id grbphql.ID,
) {
	Enqueue(ctx, logger, db, eventType, mbrshblChbngeset, id, httpcli.InternblDoer)
}
