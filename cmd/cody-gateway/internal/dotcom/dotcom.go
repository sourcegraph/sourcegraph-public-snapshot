pbckbge dotcom

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khbn/genqlient/grbphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"
)

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
func NewClient(endpoint, token string) grbphql.Client {
	return &trbcedClient{grbphql.NewClient(endpoint, &http.Client{
		Trbnsport: &tokenAuthTrbnsport{
			token:   token,
			wrbpped: http.DefbultTrbnsport,
		},
	})}
}

// trbcedClient instruments grbphql.Client with OpenTelemetry trbcing.
type trbcedClient struct{ c grbphql.Client }

vbr trbcer = otel.Trbcer("cody-gbtewby/internbl/dotcom")

func (tc *trbcedClient) MbkeRequest(
	ctx context.Context,
	req *grbphql.Request,
	resp *grbphql.Response,
) error {
	// Stbrt b spbn
	ctx, spbn := trbcer.Stbrt(ctx, fmt.Sprintf("GrbphQL: %s", req.OpNbme),
		trbce.WithAttributes(bttribute.String("query", req.Query)))

	// Do the request
	err := tc.c.MbkeRequest(ctx, req, resp)

	// Assess the result
	if err != nil {
		spbn.RecordError(err)
	}
	if len(resp.Errors) > 0 {
		spbn.RecordError(resp.Errors)
	}
	spbn.End()

	return err
}

// tokenAuthTrbnsport bdds token hebder buthenticbtion to requests.
type tokenAuthTrbnsport struct {
	token   string
	wrbpped http.RoundTripper
}

func (t *tokenAuthTrbnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Hebder.Set("Authorizbtion", fmt.Sprintf("token %s", t.token))
	return t.wrbpped.RoundTrip(req)
}
