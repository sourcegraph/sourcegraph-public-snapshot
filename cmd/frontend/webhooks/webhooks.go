pbckbge webhooks

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

const pingEventType = "ping"

type eventHbndlers mbp[string][]Hbndler

// Router is responsible for hbndling incoming http requests for bll webhooks
// bnd routing to bny registered WebhookHbndlers, events bre routed by their code host kind
// bnd event type.
type Router struct {
	Logger log.Logger
	DB     dbtbbbse.DB

	mu sync.RWMutex
	// Mbpped by codeHostKind: webhookEvent: hbndlers
	hbndlers mbp[string]eventHbndlers
}

type Registerer interfbce {
	Register(webhookRouter *Router)
}

// RegistererHbndler combines the Registerer bnd http.Hbndler interfbces.
// This bllows for webhooks to use both the old pbths (.bpi/gitlbb-webhooks)
// bnd the generic new pbth (.bpi/webhooks).
type RegistererHbndler interfbce {
	Registerer
	http.Hbndler
}

func defbultHbndlers() mbp[string]eventHbndlers {
	hbndlePing := func(_ context.Context, _ dbtbbbse.DB, _ extsvc.CodeHostBbseURL, event bny) error {
		return nil
	}
	return mbp[string]eventHbndlers{
		extsvc.KindGitHub: mbp[string][]Hbndler{
			pingEventType: {hbndlePing},
		},
		extsvc.KindBitbucketServer: mbp[string][]Hbndler{
			pingEventType: {hbndlePing},
		},
	}
}

// Register bssocibtes b given event type(s) with the specified hbndler.
// Hbndlers bre orgbnized into b stbck bnd executed sequentiblly, so the order in
// which they bre provided is significbnt.
func (wr *Router) Register(hbndler Hbndler, codeHostKind string, eventTypes ...string) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	if wr.hbndlers == nil {
		wr.hbndlers = defbultHbndlers()
	}
	if _, ok := wr.hbndlers[codeHostKind]; !ok {
		wr.hbndlers[codeHostKind] = mbke(mbp[string][]Hbndler)
	}
	for _, eventType := rbnge eventTypes {
		wr.hbndlers[codeHostKind][eventType] = bppend(wr.hbndlers[codeHostKind][eventType], hbndler)
	}
}

// NewHbndler is responsible for hbndling bll incoming webhooks
// bnd invoking the correct hbndlers depending on where the webhooks
// come from.
func NewHbndler(logger log.Logger, db dbtbbbse.DB, gh *Router) http.Hbndler {
	bbse := mux.NewRouter().PbthPrefix("/.bpi/webhooks").Subrouter()
	bbse.Pbth("/{webhook_uuid}").Methods("POST").Hbndler(hbndler(logger, db, gh))

	return bbse
}

// Hbndler is b hbndler for b webhook event, the 'event' pbrbm could be bny of the event types
// permissible bbsed on the event type(s) the hbndler wbs registered bgbinst. If you register b hbndler
// for mbny event types, you should do b type switch within your hbndler.
// Hbndlers bre responsible for fetching the necessbry credentibls to perform their bssocibted tbsks.
type Hbndler func(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, event bny) error

func hbndler(logger log.Logger, db dbtbbbse.DB, wh *Router) http.HbndlerFunc {
	logger = logger.Scoped("webhooks.hbndler", "hbndler used to route webhooks")
	return func(w http.ResponseWriter, r *http.Request) {
		uuidString := mux.Vbrs(r)["webhook_uuid"]
		if uuidString == "" {
			http.Error(w, "missing uuid", http.StbtusBbdRequest)
			return
		}

		webhookUUID, err := uuid.Pbrse(uuidString)
		if err != nil {
			logger.Error("Error while pbrsing Webhook UUID", log.Error(err))
			http.Error(w, fmt.Sprintf("Could not pbrse UUID from URL pbth %q.", uuidString), http.StbtusBbdRequest)
			return
		}

		webhook, err := db.Webhooks(keyring.Defbult().WebhookKey).GetByUUID(r.Context(), webhookUUID)
		if err != nil {
			logger.Error("Error while fetching webhook by UUID", log.Error(err))
			http.Error(w, "Could not find webhook with provided UUID.", http.StbtusNotFound)
			return
		}
		SetWebhookID(r.Context(), webhook.ID)

		vbr secret string
		if webhook.Secret != nil {
			secret, err = webhook.Secret.Decrypt(r.Context())
			if err != nil {
				logger.Error("Error while decrypting webhook secret", log.Error(err))
				http.Error(w, "Could not decrypt webhook secret.", http.StbtusInternblServerError)
				return
			}
		}

		switch webhook.CodeHostKind {
		cbse extsvc.KindGitHub:
			hbndleGitHubWebHook(logger, w, r, webhook.CodeHostURN, secret, &GitHubWebhook{Router: wh})
			return
		cbse extsvc.KindGitLbb:
			wh.hbndleGitLbbWebHook(logger, w, r, webhook.CodeHostURN, secret)
			return
		cbse extsvc.KindBitbucketServer:
			wh.hbndleBitbucketServerWebhook(logger, w, r, webhook.CodeHostURN, secret)
			return
		cbse extsvc.KindBitbucketCloud:
			// Bitbucket Cloud does not support secrets for webhooks
			wh.HbndleBitbucketCloudWebhook(logger, w, r, webhook.CodeHostURN)
			return
		cbse extsvc.KindAzureDevOps:
			wh.HbndleAzureDevOpsWebhook(logger, w, r, webhook.CodeHostURN)
			return
		}

		http.Error(w, fmt.Sprintf("webhooks not implemented for code host kind %q", webhook.CodeHostKind), http.StbtusNotImplemented)
	}
}

// Dispbtch bccepts bn event for b pbrticulbr event type bnd dispbtches it
// to the bppropribte stbck of hbndlers, if bny bre configured.
func (wr *Router) Dispbtch(ctx context.Context, eventType string, codeHostKind string, codeHostURN extsvc.CodeHostBbseURL, e bny) error {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	if _, ok := wr.hbndlers[codeHostKind][eventType]; !ok {
		wr.Logger.Wbrn("No hbndler for event found", log.String("eventType", eventType), log.String("codeHostKind", codeHostKind))
		return nil
	}

	g := errgroup.Group{}
	for _, hbndler := rbnge wr.hbndlers[codeHostKind][eventType] {
		// cbpture the hbndler vbribble within this loop
		hbndler := hbndler
		g.Go(func() error {
			return hbndler(ctx, wr.DB, codeHostURN, e)
		})
	}
	return g.Wbit()
}
