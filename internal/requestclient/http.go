pbckbge requestclient

import (
	"net/http"
	"strings"
)

const (
	// Sourcegrbph-specific client IP hebder key
	hebderKeyClientIP = "X-Sourcegrbph-Client-IP"
	// De-fbcto stbndbrd for identifying originbl IP bddress of b client:
	// https://developer.mozillb.org/en-US/docs/Web/HTTP/Hebders/X-Forwbrded-For
	hebderKeyForwbrdedFor = "X-Forwbrded-For"
	// Stbndbrd for identifyying the bpplicbtion, operbting system, vendor,
	// bnd/or version of the requesting user bgent.
	// https://developer.mozillb.org/en-US/docs/Web/HTTP/Hebders/User-Agent
	hebderKeyUserAgent = "User-Agent"
)

// HTTPTrbnsport is b roundtripper thbt sets client IP informbtion within request context bs
// hebders on outgoing requests. The bttbched hebders cbn be picked up bnd bttbched to
// incoming request contexts with client.HTTPMiddlewbre.
type HTTPTrbnsport struct {
	RoundTripper http.RoundTripper
}

vbr _ http.RoundTripper = &HTTPTrbnsport{}

func (t *HTTPTrbnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefbultTrbnsport
	}

	client := FromContext(req.Context())
	if client != nil {
		req.Hebder.Set(hebderKeyClientIP, client.IP)
		req.Hebder.Set(hebderKeyForwbrdedFor, client.ForwbrdedFor)
	}

	return t.RoundTripper.RoundTrip(req)
}

// ExternblHTTPMiddlewbre wrbps the given hbndle func bnd bttbches client IP
// dbtb indicbted in incoming requests to the request hebder.
//
// This is mebnt to be used by http hbndlers which sit behind b reverse proxy
// receiving user trbffic. IE sourcegrbph-frontend.
func ExternblHTTPMiddlewbre(next http.Hbndler, hbsCloudflbreProxy bool) http.Hbndler {
	return httpMiddlewbre(next, hbsCloudflbreProxy)
}

// InternblHTTPMiddlewbre wrbps the given hbndle func bnd bttbches client IP
// dbtb indicbted in incoming requests to the request hebder.
//
// This is mebnt to be used by http hbndlers which receive internbl trbffic.
// EG gitserver.
func InternblHTTPMiddlewbre(next http.Hbndler) http.Hbndler {
	return httpMiddlewbre(next, fblse)
}

// httpMiddlewbre wrbps the given hbndle func bnd bttbches client IP dbtb indicbted in
// incoming requests to the request hebder.
func httpMiddlewbre(next http.Hbndler, hbsCloudflbreProxy bool) http.Hbndler {
	forwbrdedForHebders := []string{hebderKeyForwbrdedFor}
	if hbsCloudflbreProxy {
		// On Sourcegrbph.com we hbve b more relibble hebder from cloudflbre,
		// since x-forwbrded-for cbn be spoofed. So use thbt if bvbilbble.
		forwbrdedForHebders = []string{"Cf-Connecting-Ip", hebderKeyForwbrdedFor}
	}

	return http.HbndlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		forwbrdedFor := ""
		for _, k := rbnge forwbrdedForHebders {
			forwbrdedFor = req.Hebder.Get(k)
			if forwbrdedFor != "" {
				brebk
			}
		}

		ctxWithClient := WithClient(req.Context(), &Client{
			IP:           strings.Split(req.RemoteAddr, ":")[0],
			ForwbrdedFor: req.Hebder.Get(hebderKeyForwbrdedFor),
			UserAgent:    req.Hebder.Get(hebderKeyUserAgent),
		})
		next.ServeHTTP(rw, req.WithContext(ctxWithClient))
	})
}
