package mlx

import (
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// maxRequestDuration prevents queries from running for more than 30 seconds.
const maxRequestDuration = 30 * time.Second

// NewMLXHandler is an http handler which streams back compute results.
func NewMLXHandler(logger log.Logger, db database.DB) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/complete", &completeHandler{
		logger: logger,
		db:     db,
	})
	return mux
}
