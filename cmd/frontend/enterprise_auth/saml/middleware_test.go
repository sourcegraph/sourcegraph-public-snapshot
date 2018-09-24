package saml

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"

	"github.com/crewjam/saml/samlidp"
)

const (
	testSAMLSPCert = `-----BEGIN CERTIFICATE-----
MIICDjCCAXegAwIBAgIJAIoFKQZTXSBZMA0GCSqGSIb3DQEBCwUAMCAxHjAcBgNV
BAMMFW15c2VydmljZS5leGFtcGxlLmNvbTAeFw0xNzEwMjAwNTUyMDZaFw0xODEw
MjAwNTUyMDZaMCAxHjAcBgNVBAMMFW15c2VydmljZS5leGFtcGxlLmNvbTCBnzAN
BgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAz15QtxhV7XdrFPWT90PLjdoo9zGB6thh
ObXsp4LewnIEs7F975md75cAoOt+0EzU+BG6lTvuTUewsYqcakUQ/5wxwUDFcG+a
qTEgXpiPg/2+CT0Y2q8oU2aauevo8GONmxgx6IA7XC8QA1GY+KUnMqqgxsfF4oPJ
ztjxowHFMRkCAwEAAaNQME4wHQYDVR0OBBYEFLO+bus3/Rgzu1r+fs/GBZ9OCs7m
MB8GA1UdIwQYMBaAFLO+bus3/Rgzu1r+fs/GBZ9OCs7mMAwGA1UdEwQFMAMBAf8w
DQYJKoZIhvcNAQELBQADgYEApXtdwA66YH8DCC4VQ+4jnjwdPKH4hqguNPzTTybY
nA9asbL6ng1w1YjQyVX/tfFIS0LiM7eE/Z6IDQqS+KVKYSAaJ3jY/NDzWk80FTsj
oCdbPeKv2gRtJqNCffDPhnsWlWvYl6ExWSPdY1Ldg82oGXfy0sMmltAa8eoWHDUz
y0M=
-----END CERTIFICATE-----`

	testSAMLSPKey = `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAM9eULcYVe13axT1
k/dDy43aKPcxgerYYTm17KeC3sJyBLOxfe+Zne+XAKDrftBM1PgRupU77k1HsLGK
nGpFEP+cMcFAxXBvmqkxIF6Yj4P9vgk9GNqvKFNmmrnr6PBjjZsYMeiAO1wvEANR
mPilJzKqoMbHxeKDyc7Y8aMBxTEZAgMBAAECgYAUUfOq3XGeIXOWzDHBqx0JO3WE
M4+9iZKNayxThdl6SF35lcz3a6A0WCGxoyH8G2tLG8Gi2gqR/BJuc1y8dSQjGtwD
PDth5mf6D2oyycBQOKLbLLAVxHgTR6tWoostesYv6SOsu3ETbjs3twmekHAMiNLZ
M3GbMMZKTGN+p/tWAQJBAPNe81V9SnGiKJw/8m6Ud7RWC9Paw+e1OowmMyIvgfxP
xqWOhad6JUZQpmPKDkNTTh15qlDYjF2Qk3u5To/PuTkCQQDaIRdqqMVfZpXQGkuj
R5YD7AYKKLRLpDfgfV0s/ZiUigVEVX++Ikkxq8Cn6kcFewoRSIPv6z0FPNtEbvrv
DJbhAkEAiEzFOzvQVZPb6qZlwEimQflu5le/ICX/hD5gpOS2h/il6FLJx+JAvgCt
L3YaRtqBBUD+ggjFlCFEeCZwOVq9AQJABB+yCKMuMBKJbIjCu1CEJojUyGZimjd9
kvHrzAjzVIOTe+o94wNU7Op5VvNX6mOcGh2L2QJSggHXh2Ctv802IQJBANOOgD8y
8JUa8LnfIMfD0PeX+MkpEfhskEr1kgRifbznHLgRTPPuftZVPdnDPH1r4hS3A3cl
cS8Ku7wC4vkfpIQ=
-----END PRIVATE KEY-----`
)

var (
	idpKey = func() crypto.PrivateKey {
		b, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0OhbMuizgtbFOfwbK7aURuXhZx6VRuAs3nNibiuifwCGz6u9
yy7bOR0P+zqN0YkjxaokqFgra7rXKCdeABmoLqCC0U+cGmLNwPOOA0PaD5q5xKhQ
4Me3rt/R9C4Ca6k3/OnkxnKwnogcsmdgs2l8liT3qVHP04Oc7Uymq2v09bGb6nPu
fOrkXS9F6mSClxHG/q59AGOWsXK1xzIRV1eu8W2SNdyeFVU1JHiQe444xLoPul5t
InWasKayFsPlJfWNc8EoU8COjNhfo/GovFTHVjh9oUR/gwEFVwifIHihRE0Hazn2
EQSLaOr2LM0TsRsQroFjmwSGgI+X2bfbMTqWOQIDAQABAoIBAFWZwDTeESBdrLcT
zHZe++cJLxE4AObn2LrWANEv5AeySYsyzjRBYObIN9IzrgTb8uJ900N/zVr5VkxH
xUa5PKbOcowd2NMfBTw5EEnaNbILLm+coHdanrNzVu59I9TFpAFoPavrNt/e2hNo
NMGPSdOkFi81LLl4xoadz/WR6O/7N2famM+0u7C2uBe+TrVwHyuqboYoidJDhO8M
w4WlY9QgAUhkPyzZqrl+VfF1aDTGVf4LJgaVevfFCas8Ws6DQX5q4QdIoV6/0vXi
B1M+aTnWjHuiIzjBMWhcYW2+I5zfwNWRXaxdlrYXRukGSdnyO+DH/FhHePJgmlkj
NInADDkCgYEA6MEQFOFSCc/ELXYWgStsrtIlJUcsLdLBsy1ocyQa2lkVUw58TouW
RciE6TjW9rp31pfQUnO2l6zOUC6LT9Jvlb9PSsyW+rvjtKB5PjJI6W0hjX41wEO6
fshFELMJd9W+Ezao2AsP2hZJ8McCF8no9e00+G4xTAyxHsNI2AFTCQcCgYEA5cWZ
JwNb4t7YeEajPt9xuYNUOQpjvQn1aGOV7KcwTx5ELP/Hzi723BxHs7GSdrLkkDmi
Gpb+mfL4wxCt0fK0i8GFQsRn5eusyq9hLqP/bmjpHoXe/1uajFbE1fZQR+2LX05N
3ATlKaH2hdfCJedFa4wf43+cl6Yhp6ZA0Yet1r8CgYEAwiu1j8W9G+RRA5/8/DtO
yrUTOfsbFws4fpLGDTA0mq0whf6Soy/96C90+d9qLaC3srUpnG9eB0CpSOjbXXbv
kdxseLkexwOR3bD2FHX8r4dUM2bzznZyEaxfOaQypN8SV5ME3l60Fbr8ajqLO288
wlTmGM5Mn+YCqOg/T7wjGmcCgYBpzNfdl/VafOROVbBbhgXWtzsz3K3aYNiIjbp+
MunStIwN8GUvcn6nEbqOaoiXcX4/TtpuxfJMLw4OvAJdtxUdeSmEee2heCijV6g3
ErrOOy6EqH3rNWHvlxChuP50cFQJuYOueO6QggyCyruSOnDDuc0BM0SGq6+5g5s7
H++S/wKBgQDIkqBtFr9UEf8d6JpkxS0RXDlhSMjkXmkQeKGFzdoJcYVFIwq8jTNB
nJrVIGs3GcBkqGic+i7rTO1YPkquv4dUuiIn+vKZVoO6b54f+oPBXd4S0BnuEqFE
rdKNuCZhiaE2XD9L/O9KP1fh5bfEcKwazQ23EvpJHBMm8BGC+/YZNw==
-----END RSA PRIVATE KEY-----`))
		k, _ := x509.ParsePKCS1PrivateKey(b.Bytes)
		return k
	}()

	idpCert = func() *x509.Certificate {
		b, _ := pem.Decode([]byte(`-----BEGIN CERTIFICATE-----
MIIDBzCCAe+gAwIBAgIJAPr/Mrlc8EGhMA0GCSqGSIb3DQEBBQUAMBoxGDAWBgNV
BAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xNTEyMjgxOTE5NDVaFw0yNTEyMjUxOTE5
NDVaMBoxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBANDoWzLos4LWxTn8Gyu2lEbl4WcelUbgLN5zYm4ron8A
hs+rvcsu2zkdD/s6jdGJI8WqJKhYK2u61ygnXgAZqC6ggtFPnBpizcDzjgND2g+a
ucSoUODHt67f0fQuAmupN/zp5MZysJ6IHLJnYLNpfJYk96lRz9ODnO1Mpqtr9PWx
m+pz7nzq5F0vRepkgpcRxv6ufQBjlrFytccyEVdXrvFtkjXcnhVVNSR4kHuOOMS6
D7pebSJ1mrCmshbD5SX1jXPBKFPAjozYX6PxqLxUx1Y4faFEf4MBBVcInyB4oURN
B2s59hEEi2jq9izNE7EbEK6BY5sEhoCPl9m32zE6ljkCAwEAAaNQME4wHQYDVR0O
BBYEFB9ZklC1Ork2zl56zg08ei7ss/+iMB8GA1UdIwQYMBaAFB9ZklC1Ork2zl56
zg08ei7ss/+iMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEFBQADggEBAAVoTSQ5
pAirw8OR9FZ1bRSuTDhY9uxzl/OL7lUmsv2cMNeCB3BRZqm3mFt+cwN8GsH6f3uv
NONIhgFpTGN5LEcXQz89zJEzB+qaHqmbFpHQl/sx2B8ezNgT/882H2IH00dXESEf
y/+1gHg2pxjGnhRBN6el/gSaDiySIMKbilDrffuvxiCfbpPN0NRRiPJhd2ay9KuL
/RxQRl1gl9cHaWiouWWba1bSBb2ZPhv2rPMUsFo98ntkGCObDX6Y1SpkqmoTbrsb
GFsTG2DLxnvr4GdN1BSr0Uu/KV3adj47WkXVPeMYQti/bQmxQB8tRFhrw80qakTL
UzreO96WzlBBMtY=
-----END CERTIFICATE-----`))
		c, _ := x509.ParseCertificate(b.Bytes)
		return c
	}()
)

// newSAMLIDPServer returns a new running SAML IDP server. It is the caller's
// responsibility to call Close().
func newSAMLIDPServer(t *testing.T) (*httptest.Server, *samlidp.Server) {
	h := http.NewServeMux()
	srv := httptest.NewServer(h)

	srvURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	idpServer, err := samlidp.New(samlidp.Options{
		URL:         *srvURL,
		Key:         idpKey,
		Certificate: idpCert,
		Store:       &samlidp.MemoryStore{},
	})
	if err != nil {
		t.Fatal(err)
	}
	h.Handle("/", idpServer)

	return srv, idpServer
}

func TestMiddleware(t *testing.T) {
	idpHTTPServer, idpServer := newSAMLIDPServer(t)
	defer idpHTTPServer.Close()

	conf.Mock(&schema.SiteConfiguration{
		AppURL:               "http://example.com",
		ExperimentalFeatures: &schema.ExperimentalFeatures{},
	})
	defer conf.Mock(nil)

	config := withConfigDefaults(&schema.SAMLAuthProvider{
		Type:                        "saml",
		IdentityProviderMetadataURL: idpServer.IDP.MetadataURL.String(),
		ServiceProviderCertificate:  testSAMLSPCert,
		ServiceProviderPrivateKey:   testSAMLSPKey,
	})

	mockGetProviderValue = &provider{config: *config}
	defer func() { mockGetProviderValue = nil }()
	auth.MockProviders = []auth.Provider{mockGetProviderValue}
	defer func() { auth.MockProviders = nil }()

	cleanup := session.ResetMockSessionStore(t)
	defer cleanup()

	providerID := providerConfigID(&mockGetProviderValue.config, true)

	// Mock user
	mockedExternalID := "testuser_id"
	const mockedUserID = 123
	auth.MockCreateOrUpdateUser = func(u db.NewUser, a db.ExternalAccountSpec) (userID int32, err error) {
		if a.ServiceType == "saml" && a.ServiceID == idpServer.IDP.MetadataURL.String() && a.ClientID == "http://example.com/.auth/saml/metadata" && a.AccountID == mockedExternalID {
			return mockedUserID, nil
		}
		return 0, fmt.Errorf("account %v not found in mock", a)
	}
	defer func() { auth.MockCreateOrUpdateUser = nil }()

	// Set up the test handler.
	authedHandler := http.NewServeMux()
	authedHandler.Handle("/.api/", Middleware.API(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid := actor.FromContext(r.Context()).UID; uid != mockedUserID && uid != 0 {
			t.Errorf("got actor UID %d, want %d", uid, mockedUserID)
		}
	})))
	authedHandler.Handle("/", Middleware.App(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte("This is the home"))
		case "/page":
			w.Write([]byte("This is a page"))
		case "/require-authn":
			actr := actor.FromContext(r.Context())
			if actr.UID == 0 {
				t.Errorf("in authn expected-endpoint, no actor was set; expected actor with UID %d", mockedUserID)
			} else if actr.UID != mockedUserID {
				t.Errorf("in authn expected-endpoint, actor with incorrect UID was set; %d != %d", actr.UID, mockedUserID)
			}
			w.Write([]byte("Authenticated"))
		default:
			http.Error(w, "", http.StatusNotFound)
		}
	})))

	// doRequest simulates a request to our authed handler (i.e., the SAML Service Provider).
	//
	// authed sets an authed actor in the request context to simulate an authenticated request.
	doRequest := func(method, urlStr, body string, cookies []*http.Cookie, authed bool, form url.Values) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		if form != nil {
			req.PostForm = form
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
		if authed {
			req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: mockedUserID}))
		}
		respRecorder := httptest.NewRecorder()
		authedHandler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	var (
		authnRequest    saml.AuthnRequest
		authnCookies    []*http.Cookie
		authnRequestURL string
	)
	t.Run("unauthenticated homepage visit -> IDP SSO URL", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/", "", nil, false, nil)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		locURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(locURL.String(), idpServer.IDP.SSOURL.String()) {
			t.Error("wrong redirect URL")
		}

		// save cookies and Authn request
		authnCookies = unexpiredCookies(resp)
		authnRequestURL = locURL.String()
		deflatedSAMLRequest, err := base64.StdEncoding.DecodeString(locURL.Query().Get("SAMLRequest"))
		if err != nil {
			t.Fatal(err)
		}
		if err := xml.NewDecoder(flate.NewReader(bytes.NewBuffer(deflatedSAMLRequest))).Decode(&authnRequest); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("unauthenticated API visit -> pass through", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.api/foo", "", nil, false, nil)
		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("wrong response code: got %v, want %v", got, want)
		}
	})
	var (
		loggedInCookies []*http.Cookie
	)
	t.Run("get SP metadata and register SP with IDP", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.auth/saml/metadata?pc="+providerID, "", nil, false, nil)
		service := samlidp.Service{}
		if err := xml.NewDecoder(resp.Body).Decode(&service.Metadata); err != nil {
			t.Fatal(err)
		}
		serviceMetadataBytes, err := xml.Marshal(service.Metadata)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", idpHTTPServer.URL+"/services/id", bytes.NewBuffer(serviceMetadataBytes))
		if err != nil {
			t.Fatal(err)
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("could not register SP with IDP, error: %s, resp: %v", err, resp)
		}
		defer resp.Body.Close()
		if want := http.StatusNoContent; resp.StatusCode != want {
			t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
		}
	})
	t.Run("get SAML assertion from IDP and post the assertion to the SP ACS URL", func(t *testing.T) {
		authnReq, err := http.NewRequest("GET", authnRequestURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		idpAuthnReq, err := saml.NewIdpAuthnRequest(&idpServer.IDP, authnReq)
		if err != nil {
			t.Fatal(err)
		}
		if err := idpAuthnReq.Validate(); err != nil {
			t.Fatal(err)
		}
		session := saml.Session{
			ID:         "session-id",
			CreateTime: time.Now(),
			ExpireTime: time.Now().Add(24 * time.Hour),
			Index:      "index",

			NameID:    "testuser_id",
			UserName:  "testuser_username",
			UserEmail: "testuser@email.com",
		}
		if err := (saml.DefaultAssertionMaker{}).MakeAssertion(idpAuthnReq, &session); err != nil {
			t.Fatal(err)
		}
		if err := idpAuthnReq.MakeResponse(); err != nil {
			t.Fatal(err)
		}
		doc := etree.NewDocument()
		doc.SetRoot(idpAuthnReq.ResponseEl)
		responseBuf, err := doc.WriteToBytes()
		if err != nil {
			t.Fatal(err)
		}
		samlResponse := base64.StdEncoding.EncodeToString(responseBuf)
		reqParams := url.Values{}
		reqParams.Set("SAMLResponse", samlResponse)
		reqParams.Set("RelayState", idpAuthnReq.RelayState)
		resp := doRequest("POST", "http://example.com/.auth/saml/acs", "", authnCookies, false, reqParams)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got status code %v, want %v", resp.StatusCode, want)
		}
		if got, want1, want2 := resp.Header.Get("Location"), "http://example.com/", "/"; got != want1 && got != want2 {
			t.Errorf("got redirect location %v, want %v or %v", got, want1, want2)
		}

		// save the cookies from the login response
		loggedInCookies = unexpiredCookies(resp)
	})
	t.Run("authenticated request to home page", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/", "", loggedInCookies, true, nil)
		respBody, _ := ioutil.ReadAll(resp.Body)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got status code %v, want %v", resp.StatusCode, want)
		}
		if got, want := string(respBody), "This is the home"; got != want {
			t.Errorf("got response body %v, want %v", got, want)
		}
	})
	t.Run("authenticated request to sub page", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/page", "", loggedInCookies, true, nil)
		respBody, _ := ioutil.ReadAll(resp.Body)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got status code %v, want %v", resp.StatusCode, want)
		}
		if got, want := string(respBody), "This is a page"; got != want {
			t.Errorf("got response body %v, want %v", got, want)
		}
	})
	t.Run("verify actor gets set in request context", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/require-authn", "", loggedInCookies, true, nil)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got status code %v, want %v", resp.StatusCode, want)
		}
	})
	t.Run("verify actor gets set in API request context", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.api/foo", "", loggedInCookies, true, nil)
		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("wrong status code: got %v, want %v", got, want)
		}
	})
}

// unexpiredCookies returns the list of unexpired cookies set by the response
func unexpiredCookies(resp *http.Response) (cookies []*http.Cookie) {
	for _, cookie := range resp.Cookies() {
		if cookie.RawExpires == "" || cookie.Expires.After(time.Now()) {
			cookies = append(cookies, cookie)
		}
	}
	return
}
