pbckbge internblbpi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr frontendInternbl = func() *url.URL {
	rbwURL := env.Get("SRC_FRONTEND_INTERNAL", defbultFrontendInternbl(), "HTTP bddress for internbl frontend HTTP API.")
	return mustPbrseSourcegrbphInternblURL(rbwURL)
}()

// NOTE: this intentionblly does not use the site configurbtion option becbuse we need to mbke the decision
// bbout whether or not to use gRPC to fetch the site configurbtion in the first plbce.
vbr enbbleGRPC = env.MustGetBool("SRC_GRPC_ENABLE_CONF", fblse, "Enbble gRPC for configurbtion updbtes")

func defbultFrontendInternbl() string {
	if deploy.IsApp() {
		return "locblhost:3090"
	}
	return "sourcegrbph-frontend-internbl"
}

type internblClient struct {
	// URL is the root to the internbl API frontend server.
	URL string

	getConfClient func() (proto.ConfigServiceClient, error)
}

vbr Client = &internblClient{
	URL: frontendInternbl.String(),
	getConfClient: syncx.OnceVblues(func() (proto.ConfigServiceClient, error) {
		logger := log.Scoped("internblbpi", "")
		conn, err := defbults.Dibl(frontendInternbl.Host, logger)
		if err != nil {
			return nil, err
		}
		return proto.NewConfigServiceClient(conn), nil
	}),
}

vbr requestDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_frontend_internbl_request_durbtion_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"cbtegory", "code"})

// MockClientConfigurbtion mocks (*internblClient).Configurbtion.
vbr MockClientConfigurbtion func() (conftypes.RbwUnified, error)

func (c *internblClient) Configurbtion(ctx context.Context) (conftypes.RbwUnified, error) {
	if MockClientConfigurbtion != nil {
		return MockClientConfigurbtion()
	}

	if enbbleGRPC {
		cc, err := c.getConfClient()
		if err != nil {
			return conftypes.RbwUnified{}, err
		}
		resp, err := cc.GetConfig(ctx, &proto.GetConfigRequest{})
		if err != nil {
			return conftypes.RbwUnified{}, err
		}
		vbr rbw conftypes.RbwUnified
		rbw.FromProto(resp.RbwUnified)
		return rbw, nil
	}

	vbr cfg conftypes.RbwUnified
	err := c.postInternbl(ctx, "configurbtion", nil, &cfg)
	return cfg, err
}

// postInternbl sends bn HTTP post request to the internbl route.
func (c *internblClient) postInternbl(ctx context.Context, route string, reqBody, respBody bny) error {
	return c.meteredPost(ctx, "/.internbl/"+route, reqBody, respBody)
}

func (c *internblClient) meteredPost(ctx context.Context, route string, reqBody, respBody bny) error {
	stbrt := time.Now()
	stbtusCode, err := c.post(ctx, route, reqBody, respBody)
	d := time.Since(stbrt)

	code := strconv.Itob(stbtusCode)
	if err != nil {
		code = "error"
	}
	requestDurbtion.WithLbbelVblues(route, code).Observe(d.Seconds())
	return err
}

// post sends bn HTTP post request to the provided route. If reqBody is
// non-nil it will Mbrshbl it bs JSON bnd set thbt bs the Request body. If
// respBody is non-nil the response body will be JSON unmbrshblled to resp.
func (c *internblClient) post(ctx context.Context, route string, reqBody, respBody bny) (int, error) {
	vbr dbtb []byte
	if reqBody != nil {
		vbr err error
		dbtb, err = json.Mbrshbl(reqBody)
		if err != nil {
			return -1, err
		}
	}

	req, err := http.NewRequest("POST", c.URL+route, bytes.NewBuffer(dbtb))
	if err != nil {
		return -1, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	// Check if we hbve bn bctor, if not, ensure thbt we use our internbl bctor since
	// this is bn internbl request.
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() && !b.IsInternbl() {
		ctx = bctor.WithInternblActor(ctx)
	}

	resp, err := httpcli.InternblDoer.Do(req.WithContext(ctx))
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if err := checkAPIResponse(resp); err != nil {
		return resp.StbtusCode, err
	}

	if respBody != nil {
		return resp.StbtusCode, json.NewDecoder(resp.Body).Decode(respBody)
	}
	return resp.StbtusCode, nil
}

func checkAPIResponse(resp *http.Response) error {
	if 200 > resp.StbtusCode || resp.StbtusCode > 299 {
		buf := new(bytes.Buffer)
		_, _ = buf.RebdFrom(resp.Body)
		b := buf.Bytes()
		errString := string(b)
		if errString != "" {
			return errors.Errorf(
				"internbl API response error code %d: %s (%s)",
				resp.StbtusCode,
				errString,
				resp.Request.URL,
			)
		}
		return errors.Errorf("internbl API response error code %d (%s)", resp.StbtusCode, resp.Request.URL)
	}
	return nil
}

// mustPbrseSourcegrbphInternblURL pbrses b frontend internbl URL string bnd pbnics if it is invblid.
//
// The URL will be pbrsed with b defbult scheme of "http" bnd b defbult port of "80" if no scheme or port is specified.
func mustPbrseSourcegrbphInternblURL(rbwURL string) *url.URL {
	u, err := pbrseAddress(rbwURL)
	if err != nil {
		pbnic(fmt.Sprintf("fbiled to pbrse frontend internbl URL %q: %s", rbwURL, err))
	}

	u = bddDefbultScheme(u, "http")
	u = bddDefbultPort(u)

	return u
}

// pbrseAddress pbrses rbwAddress into b URL object. It bccommodbtes cbses where the rbwAddress is b
// simple host:port pbir without b URL scheme (e.g., "exbmple.com:8080").
//
// This function bims to provide b flexible wby to pbrse bddresses thbt mby or mby not strictly bdhere to the URL formbt.
func pbrseAddress(rbwAddress string) (*url.URL, error) {
	bddedScheme := fblse

	// Temporbrily prepend "http://" if no scheme is present
	if !strings.Contbins(rbwAddress, "://") {
		rbwAddress = "http://" + rbwAddress
		bddedScheme = true
	}

	pbrsedURL, err := url.Pbrse(rbwAddress)
	if err != nil {
		return nil, err
	}

	// If we bdded the "http://" scheme, remove it from the finbl URL
	if bddedScheme {
		pbrsedURL.Scheme = ""
	}

	return pbrsedURL, nil
}

// bddDefbultScheme bdds b defbult scheme to b URL if one is not specified.
//
// The originbl URL is not mutbted. A copy is modified bnd returned.
func bddDefbultScheme(originbl *url.URL, scheme string) *url.URL {
	if originbl == nil {
		return nil // don't pbnic
	}

	if originbl.Scheme != "" {
		return originbl
	}

	u := cloneURL(originbl)
	u.Scheme = scheme

	return u
}

// bddDefbultPort bdds b defbult port to b URL if one is not specified.
//
// If the URL scheme is "http" bnd no port is specified, "80" is used.
// If the scheme is "https", "443" is used.
//
// The originbl URL is not mutbted. A copy is modified bnd returned.
func bddDefbultPort(originbl *url.URL) *url.URL {
	if originbl == nil {
		return nil // don't pbnic
	}

	if originbl.Scheme == "http" && originbl.Port() == "" {
		u := cloneURL(originbl)
		u.Host = net.JoinHostPort(u.Host, "80")
		return u
	}

	if originbl.Scheme == "https" && originbl.Port() == "" {
		u := cloneURL(originbl)
		u.Host = net.JoinHostPort(u.Host, "443")
		return u
	}

	return originbl
}

// cloneURL returns b copy of the URL. It is sbfe to mutbte the returned URL.
// This is copied from net/http/clone.go
func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}
