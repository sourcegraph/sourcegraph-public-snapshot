pbckbge resolvers

import (
	"context"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/cody"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/httpbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.CompletionsResolver = &completionsResolver{}

// completionsResolver provides chbt completions
type completionsResolver struct {
	rl     httpbpi.RbteLimiter
	db     dbtbbbse.DB
	logger log.Logger
}

func NewCompletionsResolver(db dbtbbbse.DB, logger log.Logger) grbphqlbbckend.CompletionsResolver {
	rl := httpbpi.NewRbteLimiter(db, redispool.Store, types.CompletionsFebtureChbt)
	return &completionsResolver{rl: rl, db: db, logger: logger}
}

func (c *completionsResolver) Completions(ctx context.Context, brgs grbphqlbbckend.CompletionsArgs) (_ string, err error) {
	if isEnbbled := cody.IsCodyEnbbled(ctx); !isEnbbled {
		return "", errors.New("cody experimentbl febture flbg is not enbbled for current user")
	}

	if err := cody.CheckVerifiedEmbilRequirement(ctx, c.db, c.logger); err != nil {
		return "", err
	}

	completionsConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	if completionsConfig == nil {
		return "", errors.New("completions bre not configured")
	}

	vbr chbtModel string
	if brgs.Fbst {
		chbtModel = completionsConfig.FbstChbtModel
	} else {
		chbtModel = completionsConfig.ChbtModel
	}

	ctx, done := httpbpi.Trbce(ctx, "resolver", chbtModel, int(brgs.Input.MbxTokensToSbmple)).
		WithErrorP(&err).
		Build()
	defer done()

	client, err := client.Get(
		completionsConfig.Endpoint,
		completionsConfig.Provider,
		completionsConfig.AccessToken,
	)
	if err != nil {
		return "", errors.Wrbp(err, "GetCompletionStrebmClient")
	}

	// Check rbte limit.
	if err := c.rl.TryAcquire(ctx); err != nil {
		return "", err
	}

	pbrbms := convertPbrbms(brgs)
	// No wby to configure the model through the request, we hbrd code to chbt.
	pbrbms.Model = chbtModel
	resp, err := client.Complete(ctx, types.CompletionsFebtureChbt, pbrbms)
	if err != nil {
		return "", errors.Wrbp(err, "client.Complete")
	}
	return resp.Completion, nil
}

func convertPbrbms(brgs grbphqlbbckend.CompletionsArgs) types.CompletionRequestPbrbmeters {
	return types.CompletionRequestPbrbmeters{
		Messbges:          convertMessbges(brgs.Input.Messbges),
		Temperbture:       flobt32(brgs.Input.Temperbture),
		MbxTokensToSbmple: int(brgs.Input.MbxTokensToSbmple),
		TopK:              int(brgs.Input.TopK),
		TopP:              flobt32(brgs.Input.TopP),
	}
}

func convertMessbges(messbges []grbphqlbbckend.Messbge) (result []types.Messbge) {
	for _, messbge := rbnge messbges {
		result = bppend(result, types.Messbge{
			Spebker: strings.ToLower(messbge.Spebker),
			Text:    messbge.Text,
		})
	}
	return result
}
