package auth

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

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

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

func Test_newSAMLAuthHandler(t *testing.T) {
	idpHTTPServer, idpServer := newSAMLIDPServer(t)
	defer idpHTTPServer.Close()

	// Mock user
	mockedUserID := samlToAuthID(idpHTTPServer.URL+"/metadata", "testuser_id")
	localstore.Mocks.Users.GetByAuthID = func(ctx context.Context, uid string) (*sourcegraph.User, error) {
		if uid == mockedUserID {
			return &sourcegraph.User{ID: 123, AuthID: uid, Username: uid}, nil
		}
		return nil, fmt.Errorf("user %q not found in mock", uid)
	}

	// Set SAML global parameters
	var err error
	samlProvider = &schema.SAMLAuthProvider{
		IdentityProviderMetadataURL: idpServer.IDP.MetadataURL.String(),
		ServiceProviderCertificate:  testSAMLSPCert,
		ServiceProviderPrivateKey:   testSAMLSPKey,
	}
	idpMetadataURL, err = url.Parse(samlProvider.IdentityProviderMetadataURL)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate an app
	appHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "":
			w.Write([]byte("This is the home"))
		case "/page":
			w.Write([]byte("This is a page"))
		case "/require-authn":
			actr := actor.FromContext(r.Context())
			if actr.UID == "" {
				t.Errorf("in authn expected-endpoint, no actor was set; expected actor with UID %q", mockedUserID)
			} else if actr.UID != mockedUserID {
				t.Errorf("in authn expected-endpoint, actor with incorrect UID was set; %q != %q", actr.UID, mockedUserID)
			}
			w.Write([]byte("Authenticated"))
		default:
			http.Error(w, "", http.StatusNotFound)
		}
	})

	authedHandler, err := newSAMLAuthHandler(context.Background(), appHandler, appURL)
	if err != nil {
		t.Fatal(err)
	}

	// doRequest simulates a request to our authed handler (i.e., the SAML Service Provider)
	doRequest := func(method, urlStr, body string, cookies []*http.Cookie, form url.Values) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		if form != nil {
			req.PostForm = form
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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
	{
		t.Logf("unauthenticated homepage visit -> IDP SSO URL")
		resp := doRequest("GET", appURL, "", nil, nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		locURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			t.Fatal(err)
		}
		check(t, strings.HasPrefix(locURL.String(), idpServer.IDP.SSOURL.String()), "wrong redirect URL")

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
	}
	var (
		loggedInCookies []*http.Cookie
	)
	{
		t.Logf("get SP metadata and register SP with IDP")
		resp := doRequest("GET", appURL+"/.auth/saml/metadata", "", nil, nil)
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
		if resp, err := http.DefaultClient.Do(req); err != nil {
			t.Fatalf("could not register SP with IDP, error: %s, resp: %v", err, resp)
		}
	}
	{
		t.Logf("get SAML assertion from IDP and post the assertion to the SP ACS URL")
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
		resp := doRequest("POST", appURL+"/.auth/saml/acs", "", authnCookies, reqParams)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong status code")
		checkEq(t, appURL, resp.Header.Get("Location"), "wrong redirect location")

		// save the cookies from the login response
		loggedInCookies = unexpiredCookies(resp)
	}
	{
		t.Logf("authenticated request to home page")
		resp := doRequest("GET", appURL, "", loggedInCookies, nil)
		respBody, _ := ioutil.ReadAll(resp.Body)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong status code")
		checkEq(t, "This is the home", string(respBody), "wrong response body")
	}
	{
		t.Logf("authenticated request to sub page")
		resp := doRequest("GET", appURL+"/page", "", loggedInCookies, nil)
		respBody, _ := ioutil.ReadAll(resp.Body)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong status code")
		checkEq(t, "This is a page", string(respBody), "wrong response body")
	}
	{
		t.Logf("verify actor gets set in request context")
		resp := doRequest("GET", appURL+"/require-authn", "", loggedInCookies, nil)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong status code")
	}
}
