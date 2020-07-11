package campaigns

import (
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type GitLabWebhook struct{ *Webhook }

func NewGitLabWebhook(store *Store, repos repos.Store, now func() time.Time) *GitLabWebhook {
	return &GitLabWebhook{&Webhook{store, repos, now, extsvc.TypeGitLab}}
}

// ServeHTTP implements the http.Handler interface.
func (h *GitLabWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implement.
	respond(w, http.StatusNotImplemented, nil)
}
