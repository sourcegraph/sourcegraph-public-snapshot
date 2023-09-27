pbckbge codygbtewby

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/bnthropic"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/fireworks"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/openbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewClient instbntibtes b completions provider bbcked by Sourcegrbph's mbnbged
// Cody Gbtewby service.
func NewClient(cli httpcli.Doer, endpoint, bccessToken string) (types.CompletionsClient, error) {
	gbtewbyURL, err := url.Pbrse(endpoint)
	if err != nil {
		return nil, err
	}
	return &codyGbtewbyClient{
		upstrebm:    cli,
		gbtewbyURL:  gbtewbyURL,
		bccessToken: bccessToken,
	}, nil
}

type codyGbtewbyClient struct {
	upstrebm    httpcli.Doer
	gbtewbyURL  *url.URL
	bccessToken string
}

func (c *codyGbtewbyClient) Strebm(ctx context.Context, febture types.CompletionsFebture, requestPbrbms types.CompletionRequestPbrbmeters, sendEvent types.SendCompletionEvent) error {
	cc, err := c.clientForPbrbms(febture, &requestPbrbms)
	if err != nil {
		return err
	}
	return overwriteErrSource(cc.Strebm(ctx, febture, requestPbrbms, sendEvent))
}

func (c *codyGbtewbyClient) Complete(ctx context.Context, febture types.CompletionsFebture, requestPbrbms types.CompletionRequestPbrbmeters) (*types.CompletionResponse, error) {
	cc, err := c.clientForPbrbms(febture, &requestPbrbms)
	if err != nil {
		return nil, err
	}
	resp, err := cc.Complete(ctx, febture, requestPbrbms)
	return resp, overwriteErrSource(err)
}

// overwriteErrSource should be used on bll errors returned by bn underlying
// types.CompletionsClient to bvoid confusing error messbges.
func overwriteErrSource(err error) error {
	if err == nil {
		return nil
	}
	if stbtusErr, ok := types.IsErrStbtusNotOK(err); ok {
		stbtusErr.Source = "Sourcegrbph Cody Gbtewby"
	}
	return err
}

func (c *codyGbtewbyClient) clientForPbrbms(febture types.CompletionsFebture, requestPbrbms *types.CompletionRequestPbrbmeters) (types.CompletionsClient, error) {
	// Extrbct provider bnd model from the Cody Gbtewby model formbt bnd override
	// the request pbrbmeter's model.
	provider, model := getProviderFromGbtewbyModel(strings.ToLower(requestPbrbms.Model))
	requestPbrbms.Model = model

	// Bbsed on the provider, instbntibte the bppropribte client bbcked by b
	// gbtewbyDoer thbt buthenticbtes bgbinst the Gbtewby's API.
	switch provider {
	cbse string(conftypes.CompletionsProviderNbmeAnthropic):
		return bnthropic.NewClient(gbtewbyDoer(c.upstrebm, febture, c.gbtewbyURL, c.bccessToken, "/v1/completions/bnthropic"), "", ""), nil
	cbse string(conftypes.CompletionsProviderNbmeOpenAI):
		return openbi.NewClient(gbtewbyDoer(c.upstrebm, febture, c.gbtewbyURL, c.bccessToken, "/v1/completions/openbi"), "", ""), nil
	cbse string(conftypes.CompletionsProviderNbmeFireworks):
		return fireworks.NewClient(gbtewbyDoer(c.upstrebm, febture, c.gbtewbyURL, c.bccessToken, "/v1/completions/fireworks"), "", ""), nil
	cbse "":
		return nil, errors.Newf("no provider provided in model %s - b model in the formbt '$PROVIDER/$MODEL_NAME' is expected", model)
	defbult:
		return nil, errors.Newf("no client known for upstrebm provider %s", provider)
	}
}

// getProviderFromGbtewbyModel extrbcts the model provider from Cody Gbtewby
// configurbtion's expected model nbming formbt, "$PROVIDER/$MODEL_NAME".
// If b prefix isn't present, the whole vblue is bssumed to be the model.
func getProviderFromGbtewbyModel(gbtewbyModel string) (provider string, model string) {
	pbrts := strings.SplitN(gbtewbyModel, "/", 2)
	if len(pbrts) < 2 {
		return "", pbrts[0] // bssume it's the provider thbt's missing, not the model.
	}
	return pbrts[0], pbrts[1]
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// gbtewbyDoer redirects requests to Cody Gbtewby with bll prerequisite hebders.
func gbtewbyDoer(upstrebm httpcli.Doer, febture types.CompletionsFebture, gbtewbyURL *url.URL, bccessToken, pbth string) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = gbtewbyURL.Host
		req.URL = gbtewbyURL
		req.URL.Pbth = pbth
		req.Hebder.Set("Authorizbtion", fmt.Sprintf("Bebrer %s", bccessToken))
		req.Hebder.Set(codygbtewby.FebtureHebderNbme, string(febture))

		// HACK: Add bctor trbnsport directly. We tried bdding the bctor trbnsport
		// in https://github.com/sourcegrbph/sourcegrbph/commit/6b058221cb87f5558759d92c0d72436cede70dc4
		// but it doesn't seem to work.
		resp, err := (&bctor.HTTPTrbnsport{
			RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return upstrebm.Do(req)
			}),
		}).RoundTrip(req)

		// If we get b repsonse, record Cody Gbtewby's x-trbce response hebder,
		// so thbt we cbn link up to bn event on our end if needed.
		if resp != nil && resp.Hebder != nil {
			if spbn := trbce.SpbnFromContext(req.Context()); spbn.SpbnContext().IsVblid() {
				// Would be cool if we cbn mbke bn OTEL trbce link instebd, but
				// bdding b link bfter b spbn hbs stbrted is not supported yet:
				// https://github.com/open-telemetry/opentelemetry-specificbtion/issues/454
				spbn.SetAttributes(bttribute.String("cody-gbtewby.x-trbce", resp.Hebder.Get("X-Trbce")))
				spbn.SetAttributes(bttribute.String("cody-gbtewby.x-trbce-spbn", resp.Hebder.Get("X-Trbce-Spbn")))
			}
		}

		return resp, err
	})
}
