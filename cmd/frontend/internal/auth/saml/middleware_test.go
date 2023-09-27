pbckbge sbml

import (
	"bytes"
	"compress/flbte"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/bbse64"
	"encoding/pem"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjbm/sbml"
	"github.com/crewjbm/sbml/sbmlidp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	testSAMLSPCert = `-----BEGIN CERTIFICATE-----
MIIC+zCCAeOgAwIBAgIQFkK4RCQNFkAFzj8dJHnXJjANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTE4MTAyMDA4NTQ1NVoXDTI4MTAxNzA4NTQ1
NVowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBAKMD2RmTnAI+1+s+hibkkdbOXHoEwRoG45yeCV5z8A7TnZtF238kReBN
JSOUTvgrvg5WbfG8ULSbriepAI45BH3yYoNOBXe0biVsCB+0h6szeV1+N6y9wj0j
ns/AOOV6ec/GbUZufF+XeJmVX/kRoOthUCEWhCGn/ZCb9VNcr2u/EhCZhvk6JcY9
p/gu2YYJepihYpkrzzHwlC+ye+AfPX0/LiZQLGM8ciiziXden8DqEhskkg5HqnPl
hwscqI6qlYIcUFw5QB3xA738N4/92Uj7Jstf05ESFDf6zbUTn/hSsLXivNHI0G4P
4gsVy5Y5pygrw3b3FuodJbuVtLU9cwMCAwEAAbNNMEswDgYDVR0PAQH/BAQDAgWg
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFgYDVR0RBA8wDYIL
ZXhhbXBsZS5jb20wDQYJKoZIhvcNAQELBQADggEBADQ/UgbXlW7zPwWswJSlbgph
yjepD5dJ/My1ByIM2GSSYlvnLGq9tSOwUWZ0fZY/G8WOowNSBQlUVPT7tS11j7Ce
BdrImHWpDleZYybgk08vbU059LJFI4EM8Vzn570h3dxdXkSoIGjqNAfywbt591k3
K8llPk2wrQ8fv3KA7tNNmJW+Ee1DHIAui53bFe3sHmp7JN1tE6HlqrSLIDymSd28
tOfJ1Y9kOvUF7DY8pkSVDukO9wsy0X568hfJOz4PQe/1LHJ1YxlomTCkyVV8xtW7
hbnEyvPo2yr/SHbk4Fz1yXP20cBm8vO2DmKlI0kbKGQw1Rybl8NQw+OPdb/V6pM=
-----END CERTIFICATE-----`

	testSAMLSPKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAowPZGZOcAj7X6z6GJqSR1s5cegTBGgbjnJ4JXnPwDtOdm0Xb
fyRF4E0lI5RO+Cu+DlZt8bxQtJquJ6kAjjkEffJig04Fd7RuJWwIH7SHqzN5XX43
rL3CPSOez8A45Xp5z8ZtRm58X5d4mZVf+RGg62FQIRbEIbf9kJr1U1yvb78SEJmG
+Tolxj2n+C7Zhgl6mKFimSvPMfCUL7J74B89fT8uJlAsYzxyKLOJd16fwOoSGySS
Dkeqc+WHCxyojqqVghxQXDlAHfEDvfw3j/3ZSPsmy1/TkRIUN/rNtROf+FKwteK8
0cjQbg/iCxXLljmnKCvDdvcW6h0lu5W0tT1zAwIDAQABAoIBAC/t4LYhbVxHp+p1
zrGr72lN8Wi63x/M6L1SxgRsbCej1pIhvwCp5JWneQT2BSX4jn/er6LEsKH5XL0y
doRbhVSWoJpkpTzl4wDDu7u+s6kFkGiJxMrYXDTntTj2FoR6Nzh86gIsWAsvGPln
LvmnUj4CtbGU0jKnFumedgUVmko+QmblghYrkc7dReprGJ6EDWNLGb0ASG9/R7iQ
NOKu17nK9W3yCWJc48SR8y9HWUEUqtKbsshJ6PewNNttsSC3JjeGuiH5fRmXLi6L
wXr2l1AAPGRWbI8djrm7DFLb1s8pfJKkTV0YUDHFNBXny0h+oUGwC+KCQNsfE3t8
GbKqdKkCgYEA0A3dFKuxzm9QbZRcmGwZbHTNfWb7EMlknNq9wqZLgbTK77P1bhXW
l0YP7HuNZnKToMt3UrM5tYFzk28b73p4Ur8+vb23lbtwwBeZ5qsZ/vAfI9b2GTj5
AchOI5Xtwf8eWT3OIdHbe4W5hkyb/siPdkRJ1zXfDn8XN0+XIB7NeT0CgYEAyJTn
xjtxMq3Jbb5tycZ7ZVC9Y5tpd6phb6KLdLNGNcNKSPvC2EichgLH3Awckpe92HvD
wujlQnlKod42/bPxVcOOnwqN2PMfzPXK91pcOt9QyWFzRLFtvkFB9Dd4vz5uhdWF
CpSBDN87PpFEOuJJApy78e7hfZ5YpxWGK7N24T8CgYEAxtT4+9A6VTc8ffzToTdt
9KCL4dSRDDHr3ZuOzn9umb7WUs6BN3vXYSqr/Sz2rXnCbGEG4Bo4hKX6dmQwMb2x
UCNFKrDiSk6gKnRjuHb8mU+R8wZ0mxY/otxzEL8wQb42msLeRKPyRdI+w4JjctLp
h/UrPGlXitsbrNl7bE8Dv2ECgYBXOgog9rCfbVvtlFbCLMJ0qMvziR4wX/PHbFRh
B6U8tBSV8IYnMEyBKqxnUQ0L4tk4T3ouRMGOStjd05jubGEC/uwC1cAh3Hiz1R/S
uYTqRTsImExcTxx+ZDqeTZFA+ZFuuhAFLdeBFYLbDqoxQT6m2CoTZ+K/kiDTbFTU
pFLKWQKBgQCNCeMVMkNwJtMeR/vKUEOPZLKFZihPrORi8F5v0f+qrcg3bKDKd1KI
6kocCulZR3uvEFHUAoNyMNwCZs6YyIK9zEN3/Pb46ThnJNNMXv0CG+W8df/Vbnd0
REijBJT1tjS4dXBULokRuI2640iWll8KX1KQDzqo3l++JRGqoP57Sg==
-----END RSA PRIVATE KEY-----`
)

vbr (
	idpCert = func() *x509.Certificbte {
		b, _ := pem.Decode([]byte(`-----BEGIN CERTIFICATE-----
MIIC/DCCAeSgAwIBAgIRANzQVAHz24KrifhcM9kbVcYwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xODEwMjAwODU1NDNbFw0yODEwMTcwODU1
NDNbMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCYO2FXgBvsrHyDzFFjUxNuF5OtPYOw+yjGnMR7r1CU93ZbSFN0z9Ux
Qv6rHAmoGwLp+dJjYZ2g/km6/TnONVGjSDh8TxCEH+cP5kIyRN4L3MPW9tsZ36Fd
5yqNbVCrxp3gJKAHmcUHYONzQ6WxxOCEBkvVknysstG8hXhbOcElXrSIyRVPQuEu
TBQSAJbhFbQYCKFU93rlO142hgPJDkHibz8PhLgEU7v3Eo23JrOSKNUysXnp5hLT
RhOyQyWNpXA91wwsTwETOD2KlDKDHIcpKEdMSWhRQyb6S2z49RVKYpZmATW4Qq2l
hDWWuyG7/IbheuyPum+BF0FFmDKfQY83AgMBAAGjTTBLMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMBYGA1UdEQQPMA2C
C2V4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQBvdi9Gz1qADI0F/oC/7fh/
TjevChAO3XsuGz53yqeM0z49yHJooCV3jzgBrjX82DAhk/nQ0kF+YZummoPG1fqf
rycmsq0yD7Gy/do8sW7XSvpQpkbiBb9yg7rP1eVEfn+vDv0ZS0F3mqbfXl1v7FQ0
PwPtE49ObO7rb3FbLPBtEXocvGjvgb8SRmuT3/5oCI46AKldENL6+CKEEWWUIuW8
HeKPQIRYzcqi1dy88nRk44DkCyNxe/h7X/MGt4Mu9HjDH7lDJs0038sdghsX74ET
gN6hZwBR6U2UoJHDytj/+KtSL/XGZSwTgOFyFyMZROcqUPWRwl7Zk3dOKy+3T2Bi
-----END CERTIFICATE-----
`))
		c, _ := x509.PbrseCertificbte(b.Bytes)
		return c
	}()

	idpKey = func() crypto.PrivbteKey {
		b, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAmDthV4Ab7Kx8g8xRY1MTbheTrT2DsPsoxpzEe69QlPd2WkhT
dM/VMUL+qxwJqBsC6fnSY2GdoP5Juv05zjVRo0g4fE8QhB/nD+ZCMkTeC9zD1vbb
Gd+hXecqjW1Qq8bd4CSgB5nFB2Djc0OlscTghAZL1ZJ8rLLRvIV4WznBJV60iMkV
T0LhLkwUEgCWoRW0GAihVPd65TteNoYDyQ5B4m8/D4S4BFO79xKNtybzkijVMrF5
6eYS00YTskMljbVwPdcMLE8BEzg9ipQygxyHKShHTEloUUMmukts+PUVSmKWZgE1
uEKtpYQ1lrshu/yGoXrsj7pvgRdBRZgyn0GPNwIDAQABAoIBAQCXS7zM5+fY6vy9
SJ1C59gRvKDqto5hoNy/uCKXAoBF7UPVKri3Cb/Ky9irWqxGRMI6pC1y1BuDW/cP
Pojq5pcCfs6UzUeO6N4OMTxtFYDRrVF+Hc1YA6gu2YbzFIfukPFrSTs7Epp9YM/t
SLgu24p/7HoGAxbh1P8bLFSX5eiOJ+8t8byYOrKLp3Rn67lC9Y+9LX4X6GHlBMDc
WHYupi3ZA7Q59dXQCJHFNG/hk17AMtB8lFrb9rUid8teX8ZJKJQ26hU2O0UMujjM
mFlCdmvc97lJ4LhjrWHv/9ybcf90bViHIkL52Yux1jNt/jl3/7CyBwHbbu4b0qoZ
QkM4WIihAoGBAMlzsUeJxBCbUyMd/6TiLn60SDn0HMf4ZVdGqqxkhopVmsvRTn+P
wu9YHWFPwXwVL3wdtuBjsEA9nMKWWMQKbQUZhm1Y+AQIVpVNQqesgyLctVoIUBNY
fglvKrs8JuRuwMpE2P/3lXMsxtV9AyCpxxXhyb8KqJb2jcMB/Lr+lx+fAoGBAMFz
16yHU+Zo6OOvy7T2xh67btwOrS2kIzhGO6gcK7qbbkGGebKLShnMYEmFGp4EbHTf
OVie+SU0OWF/e5bgFWC+fm6jWyhO0xPRbvs2P+l2KtnT2UBT9IgjhrVUIzp+Vn7t
cjfb32m7km1kZZ48ySP9cH/4/xnT6XEC33PoNwlpAoGAG1t+w7xNyAOP8sDsKrQc
pFBPTq98CRwOhx+tpeOw8bBWbT9vbZtUWbSZqNFv8S3fWPegEjD3ioHTfAl23Iid
7Ydd3hOq+sE3IOdxGdwvotheOG/QkBAAbb+PCgZNMdBolg9reLdisFVwWyWy+wiT
ZMFY5lCIPI9mCQmIDMzuMPkCgYBFJKJxh+z07YpP1wV4KLunQFbfUF+VcJUmB/RK
ocb/bzL9OJNBBYf2sJW5sVlSIUE0hJR6mFd0dLYNowMJbg46Bdwqrzhlr8bBzplc
MIenbhTmxlFgLKG6Bvie1vPAdGd19mhcjrnLkL9FWhz38cHymyMbmmSTVqqZOe2j
/9usAQKBgQCT//j6XflAr20gb+mNcoJqVxRTFtSsZb23kJnZ3Sfs3R8XXu5NcZEZ
ODI9ZpZ9tg8oK9EB5vqMFTTnyjpbr7F2jqFWtUmNju/rGlrQCZx0we+EAW/R2hFP
YGYu4Z+SyXTsv/Ys5VGWuuCJO32RuRBeC4eJCmpyH0mqPhIBZmV4Jw==
-----END RSA PRIVATE KEY-----
`))
		k, _ := x509.PbrsePKCS1PrivbteKey(b.Bytes)
		return k
	}()
)

// newSAMLIDPServer returns b new running SAML IDP server. It is the cbller's
// responsibility to cbll Close().
func newSAMLIDPServer(t *testing.T) (*httptest.Server, *sbmlidp.Server) {
	h := http.NewServeMux()
	srv := httptest.NewServer(h)

	srvURL, err := url.Pbrse(srv.URL)
	if err != nil {
		t.Fbtbl(err)
	}

	idpServer, err := sbmlidp.New(sbmlidp.Options{
		URL:         *srvURL,
		Key:         idpKey,
		Certificbte: idpCert,
		Store:       &sbmlidp.MemoryStore{},
	})
	if err != nil {
		t.Fbtbl(err)
	}
	h.Hbndle("/", idpServer)

	return srv, idpServer
}

func TestMiddlewbre(t *testing.T) {
	idpHTTPServer, idpServer := newSAMLIDPServer(t)
	defer idpHTTPServer.Close()

	defer licensing.TestingSkipFebtureChecks()()

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{},
			ExternblURL:          "http://exbmple.com",
		},
	})
	defer conf.Mock(nil)

	config := withConfigDefbults(&schemb.SAMLAuthProvider{
		Type:                        "sbml",
		IdentityProviderMetbdbtbURL: idpServer.IDP.MetbdbtbURL.String(),
		ServiceProviderCertificbte:  testSAMLSPCert,
		ServiceProviderPrivbteKey:   testSAMLSPKey,
	})

	mockGetProviderVblue = &provider{config: *config}
	defer func() { mockGetProviderVblue = nil }()
	providers.MockProviders = []providers.Provider{mockGetProviderVblue}
	defer func() { providers.MockProviders = nil }()

	clebnup := session.ResetMockSessionStore(t)
	defer clebnup()

	providerID := providerConfigID(&mockGetProviderVblue.config, true)

	// Mock user
	mockedExternblID := "testuser_id"
	const mockedUserID = 123
	buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
		if op.ExternblAccount.ServiceType == "sbml" && op.ExternblAccount.ServiceID == idpServer.IDP.MetbdbtbURL.String() && op.ExternblAccount.ClientID == "http://exbmple.com/.buth/sbml/metbdbtb" && op.ExternblAccount.AccountID == mockedExternblID {
			return mockedUserID, "", nil
		}
		return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
	}
	defer func() { buth.MockGetAndSbveUser = nil }()

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, CrebtedAt: time.Now()}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Set up the test hbndler.
	buthedHbndler := http.NewServeMux()
	buthedHbndler.Hbndle("/.bpi/", Middlewbre(db).API(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid := bctor.FromContext(r.Context()).UID; uid != mockedUserID && uid != 0 {
			t.Errorf("got bctor UID %d, wbnt %d", uid, mockedUserID)
		}
	})))
	buthedHbndler.Hbndle("/", Middlewbre(db).App(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Pbth {
		cbse "/":
			_, _ = w.Write([]byte("This is the home"))
		cbse "/pbge":
			_, _ = w.Write([]byte("This is b pbge"))
		cbse "/require-buthn":
			bctr := bctor.FromContext(r.Context())
			if bctr.UID == 0 {
				t.Errorf("in buthn expected-endpoint, no bctor wbs set; expected bctor with UID %d", mockedUserID)
			} else if bctr.UID != mockedUserID {
				t.Errorf("in buthn expected-endpoint, bctor with incorrect UID wbs set; %d != %d", bctr.UID, mockedUserID)
			}
			_, _ = w.Write([]byte("Authenticbted"))
		defbult:
			http.Error(w, "", http.StbtusNotFound)
		}
	})))

	// doRequest simulbtes b request to our buthed hbndler (i.e., the SAML Service Provider).
	//
	// buthed sets bn buthed bctor in the request context to simulbte bn buthenticbted request.
	doRequest := func(method, urlStr, body string, cookies []*http.Cookie, buthed bool, form url.Vblues) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := rbnge cookies {
			req.AddCookie(cookie)
		}
		if form != nil {
			req.PostForm = form
			req.Hebder.Add("Content-Type", "bpplicbtion/x-www-form-urlencoded")
		}
		if buthed {
			req = req.WithContext(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: mockedUserID}))
		}
		respRecorder := httptest.NewRecorder()
		buthedHbndler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	vbr (
		buthnRequest    sbml.AuthnRequest
		buthnCookies    []*http.Cookie
		buthnRequestURL string
	)
	t.Run("unbuthenticbted homepbge visit, no sign-out cookie -> IDP SSO URL", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/", "", nil, fblse, nil)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		locURL, err := url.Pbrse(resp.Hebder.Get("Locbtion"))
		if err != nil {
			t.Fbtbl(err)
		}
		if !strings.HbsPrefix(locURL.String(), idpServer.IDP.SSOURL.String()) {
			t.Error("wrong redirect URL")
		}

		// sbve cookies bnd Authn request
		buthnCookies = unexpiredCookies(resp)
		buthnRequestURL = locURL.String()
		deflbtedSAMLRequest, err := bbse64.StdEncoding.DecodeString(locURL.Query().Get("SAMLRequest"))
		if err != nil {
			t.Fbtbl(err)
		}
		if err := xml.NewDecoder(flbte.NewRebder(bytes.NewBuffer(deflbtedSAMLRequest))).Decode(&buthnRequest); err != nil {
			t.Fbtbl(err)
		}
	})
	t.Run("unbuthenticbted homepbge visit, sign-out cookie present -> sg login", func(t *testing.T) {
		cookie := &http.Cookie{Nbme: buth.SignOutCookie, Vblue: "true"}

		resp := doRequest("GET", "http://exbmple.com/", "", []*http.Cookie{cookie}, fblse, nil)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("unbuthenticbted API visit -> pbss through", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.bpi/foo", "", nil, fblse, nil)
		if got, wbnt := resp.StbtusCode, http.StbtusOK; got != wbnt {
			t.Errorf("wrong response code: got %v, wbnt %v", got, wbnt)
		}
	})

	vbr loggedInCookies []*http.Cookie

	t.Run("get SP metbdbtb bnd register SP with IDP", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.buth/sbml/metbdbtb?pc="+providerID, "", nil, fblse, nil)
		service := sbmlidp.Service{}
		if err := xml.NewDecoder(resp.Body).Decode(&service.Metbdbtb); err != nil {
			t.Fbtbl(err)
		}
		serviceMetbdbtbBytes, err := xml.Mbrshbl(service.Metbdbtb)
		if err != nil {
			t.Fbtbl(err)
		}
		req, err := http.NewRequest("PUT", idpHTTPServer.URL+"/services/id", bytes.NewBuffer(serviceMetbdbtbBytes))
		if err != nil {
			t.Fbtbl(err)
		}
		resp, err = http.DefbultClient.Do(req)
		if err != nil {
			t.Fbtblf("could not register SP with IDP, error: %s, resp: %v", err, resp)
		}
		defer resp.Body.Close()
		if wbnt := http.StbtusNoContent; resp.StbtusCode != wbnt {
			t.Errorf("got HTTP %d, wbnt %d", resp.StbtusCode, wbnt)
		}
	})

	t.Run("get SAML bssertion from IDP bnd post the bssertion to the SP ACS URL", func(t *testing.T) {
		buthnReq, err := http.NewRequest("GET", buthnRequestURL, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		idpAuthnReq, err := sbml.NewIdpAuthnRequest(&idpServer.IDP, buthnReq)
		if err != nil {
			t.Fbtbl(err)
		}
		if err := idpAuthnReq.Vblidbte(); err != nil {
			t.Fbtbl(err)
		}
		sbmlSession := sbml.Session{
			ID:         "session-id",
			CrebteTime: time.Now(),
			ExpireTime: time.Now().Add(24 * time.Hour),
			Index:      "index",

			NbmeID:    "testuser_id",
			UserNbme:  "testuser_usernbme",
			UserEmbil: "testuser@embil.com",
		}
		if err := (sbml.DefbultAssertionMbker{}).MbkeAssertion(idpAuthnReq, &sbmlSession); err != nil {
			t.Fbtbl(err)
		}
		if err := idpAuthnReq.MbkeResponse(); err != nil {
			t.Fbtbl(err)
		}
		doc := etree.NewDocument()
		doc.SetRoot(idpAuthnReq.ResponseEl)
		responseBuf, err := doc.WriteToBytes()
		if err != nil {
			t.Fbtbl(err)
		}
		sbmlResponse := bbse64.StdEncoding.EncodeToString(responseBuf)
		reqPbrbms := url.Vblues{}
		reqPbrbms.Set("SAMLResponse", sbmlResponse)
		reqPbrbms.Set("RelbyStbte", idpAuthnReq.RelbyStbte)
		resp := doRequest("POST", "http://exbmple.com/.buth/sbml/bcs", "", buthnCookies, fblse, reqPbrbms)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt1, wbnt2 := resp.Hebder.Get("Locbtion"), "http://exbmple.com/", "/"; got != wbnt1 && got != wbnt2 {
			t.Errorf("got redirect locbtion %v, wbnt %v or %v", got, wbnt1, wbnt2)
		}

		// sbve the cookies from the login response
		loggedInCookies = unexpiredCookies(resp)
	})
	t.Run("buthenticbted request to home pbge", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/", "", loggedInCookies, true, nil)
		respBody, _ := io.RebdAll(resp.Body)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := string(respBody), "This is the home"; got != wbnt {
			t.Errorf("got response body %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("buthenticbted request to sub pbge", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/pbge", "", loggedInCookies, true, nil)
		respBody, _ := io.RebdAll(resp.Body)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := string(respBody), "This is b pbge"; got != wbnt {
			t.Errorf("got response body %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("verify bctor gets set in request context", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/require-buthn", "", loggedInCookies, true, nil)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("verify bctor gets set in API request context", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.bpi/foo", "", loggedInCookies, true, nil)
		if got, wbnt := resp.StbtusCode, http.StbtusOK; got != wbnt {
			t.Errorf("wrong stbtus code: got %v, wbnt %v", got, wbnt)
		}
	})
}

func TestAllowSignin(t *testing.T) {
	bllowedGroups := []string{"foo"}
	providerConfig := schemb.SAMLAuthProvider{
		AllowGroups: bllowedGroups,
	}
	mockProvider := &provider{
		config: providerConfig,
	}

	t.Run("Sign in is bllowed if bllowGroups is not configured", func(t *testing.T) {
		p := &provider{
			config: schemb.SAMLAuthProvider{},
		}
		result := bllowSignin(p, mbke(mbp[string]bool))
		if !result {
			t.Errorf("Expected bllowSigning to be true, got %v", result)
		}
	})
	t.Run("Sign in is bllowed if user belongs to b group", func(t *testing.T) {
		groups := mbke(mbp[string]bool)
		groups["foo"] = true
		result := bllowSignin(mockProvider, groups)
		if !result {
			t.Errorf("Expected bllowSigning to be true, got %v", result)
		}
	})
	t.Run("Sign in is not bllowed if user does not belong to bny group in bllowGroups", func(t *testing.T) {
		groups := mbke(mbp[string]bool)
		groups["bbr"] = true
		groups["bbz"] = true
		result := bllowSignin(mockProvider, groups)
		if result {
			t.Errorf("Expected bllowSigning to be fblse, got %v", result)
		}
	})
	t.Run("Sign in is not bllowed if bllowGroups is empty", func(t *testing.T) {
		p := &provider{
			config: schemb.SAMLAuthProvider{
				AllowGroups: []string{},
			},
		}
		result := bllowSignin(p, mbke(mbp[string]bool))
		if result {
			t.Errorf("Expected bllowSigning to be fblse, got %v", result)
		}
	})
	t.Run("Sign in is not bllowed if user groups bre empty", func(t *testing.T) {
		result := bllowSignin(mockProvider, mbke(mbp[string]bool))
		if result {
			t.Errorf("Expected bllowSigning to be fblse, got %v", result)
		}
	})
}

// unexpiredCookies returns the list of unexpired cookies set by the response
func unexpiredCookies(resp *http.Response) (cookies []*http.Cookie) {
	for _, cookie := rbnge resp.Cookies() {
		if cookie.RbwExpires == "" || cookie.Expires.After(time.Now()) {
			cookies = bppend(cookies, cookie)
		}
	}
	return
}
