package webhooks

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketCloudWebhook struct {
	*Webhook
}

func NewBitbucketCloudWebhook(store *store.Store) *BitbucketCloudWebhook {
	return &BitbucketCloudWebhook{
		Webhook: &Webhook{store, extsvc.TypeBitbucketCloud},
	}
}

func (h *BitbucketCloudWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.parseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	log15.Warn("webhook", "e", fmt.Sprintf("%+v", e), "extSvc", fmt.Sprintf("%+v", extSvc))

	respond(w, http.StatusTeapot, errors.New("🫖"))
}

func (h *BitbucketCloudWebhook) parseEvent(r *http.Request) (interface{}, *types.ExternalService, *httpError) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	rawID := r.FormValue(extsvc.IDParam)
	var externalServiceID int64
	// id could be blank temporarily if we haven't updated the hook url to include the param yet
	if rawID != "" {
		externalServiceID, err = strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "invalid external service id")}
		}
	}

	args := database.ExternalServicesListOptions{Kinds: []string{extsvc.KindBitbucketCloud}}
	if externalServiceID != 0 {
		args.IDs = append(args.IDs, externalServiceID)
	}
	es, err := h.Store.ExternalServices().List(r.Context(), args)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	var extSvc *types.ExternalService
	for _, e := range es {
		if externalServiceID != 0 && e.ID != externalServiceID {
			continue
		}

		c, _ := e.Configuration()
		con, ok := c.(*schema.BitbucketCloudConnection)
		if !ok {
			continue
		}

		if secret := con.WebhookSecret; secret != "" {
			if r.FormValue("secret") == secret {
				extSvc = e
				break
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StatusUnauthorized, err}
	}

	return payload, extSvc, nil
}
