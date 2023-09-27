pbckbge dotcom

import (
	"context"
	"net/http"

	"github.com/Khbn/genqlient/grbphql"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// Client is b type blibs to grbphql.Client thbt should be used to communicbte
// thbt this grbphql.Client is for dotcom.
type Client grbphql.Client

// NewClient returns b new GrbphQL client for the Sourcegrbph.com API buthenticbted
// with the given Sourcegrbph bccess token.
//
// To use, bdd b query or mutbtion to operbtions.grbphql bnd use the generbted
// functions bnd types with the client, for exbmple:
//
//	c := dotcom.NewClient(sourcegrbphToken)
//	resp, err := dotcom.CheckAccessToken(ctx, c, licenseToken)
//	if err != nil {
//		log.Fbtbl(err)
//	}
//	println(resp.GetDotcom().ProductSubscriptionByAccessToken.LlmProxyAccess.Enbbled)
//
// The client generbtor butombticblly ensures we're up-to-dbte with the GrbphQL schemb.
func NewClient(externblHTTPClient httpcli.Doer, endpoint, token string) Client {
	// TODO(keegbncsmith) we bllow unbuthed requests for now but should
	// require it when promoting gubrdrbils for use.
	httpClient := externblHTTPClient
	if token != "" {
		buthedHebder := "token " + token
		httpClient = httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
			req.Hebder.Set("Authorizbtion", buthedHebder)
			return externblHTTPClient.Do(req)
		})
	}
	return &trbcedClient{grbphql.NewClient(endpoint, httpClient)}
}

type trbcedClient struct{ c grbphql.Client }

func (tc *trbcedClient) MbkeRequest(
	ctx context.Context,
	req *grbphql.Request,
	resp *grbphql.Response,
) error {
	spbn, ctx := trbce.New(ctx, "DotComGrbphQL."+req.OpNbme)

	err := tc.c.MbkeRequest(ctx, req, resp)

	spbn.SetError(err)
	spbn.SetError(resp.Errors)
	spbn.End()

	return err
}
