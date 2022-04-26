package webhooks

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	respond(w, http.StatusTeapot, errors.New("🫖"))
}
