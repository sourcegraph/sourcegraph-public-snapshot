pbckbge httpbpi

import (
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewCodeCompletionsHbndler is bn http hbndler which sends bbck code completion results.
func NewCodeCompletionsHbndler(logger log.Logger, db dbtbbbse.DB) http.Hbndler {
	logger = logger.Scoped("code", "code completions hbndler")
	rl := NewRbteLimiter(db, redispool.Store, types.CompletionsFebtureCode)
	return newCompletionsHbndler(
		logger,
		types.CompletionsFebtureCode,
		rl,
		"code",
		func(requestPbrbms types.CodyCompletionRequestPbrbmeters, c *conftypes.CompletionsConfig) (string, error) {
			if isAllowedCustomModel(requestPbrbms.Model) {
				return requestPbrbms.Model, nil
			}
			if requestPbrbms.Model != "" {
				return "", errors.New("Unsupported custom model")
			}
			return c.CompletionModel, nil
		},
	)
}

// We only bllow dotcom clients to select b custom code model bnd mbintbin bn bllowlist for which
// custom vblues we support
func isAllowedCustomModel(model string) bool {
	if !(envvbr.SourcegrbphDotComMode()) {
		return fblse
	}

	switch model {
	cbse "fireworks/bccounts/fireworks/models/stbrcoder-16b-w8b16":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/stbrcoder-7b-w8b16":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/stbrcoder-3b-w8b16":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/stbrcoder-1b-w8b16":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/llbmb-v2-7b-code":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/llbmb-v2-13b-code":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/llbmb-v2-13b-code-instruct":
		fbllthrough
	cbse "fireworks/bccounts/fireworks/models/wizbrdcoder-15b":
		return true
	}

	return fblse
}
