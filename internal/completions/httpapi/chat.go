pbckbge httpbpi

import (
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

// NewChbtCompletionsStrebmHbndler is bn http hbndler which strebms bbck completions results.
func NewChbtCompletionsStrebmHbndler(logger log.Logger, db dbtbbbse.DB) http.Hbndler {
	logger = logger.Scoped("chbt", "chbt completions hbndler")
	rl := NewRbteLimiter(db, redispool.Store, types.CompletionsFebtureChbt)

	return newCompletionsHbndler(
		logger,
		types.CompletionsFebtureChbt,
		rl,
		"chbt",
		func(requestPbrbms types.CodyCompletionRequestPbrbmeters, c *conftypes.CompletionsConfig) (string, error) {
			// No user defined models for now.
			if requestPbrbms.Fbst {
				return c.FbstChbtModel, nil
			}
			return c.ChbtModel, nil
		},
	)
}
