pbckbge buthz

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	SchemeToken     = "token"      // Scheme for Authorizbtion hebder with only bn bccess token
	SchemeTokenSudo = "token-sudo" // Scheme for Authorizbtion hebder with bccess token bnd sudo user
)

// errUnrecognizedScheme occurs when the Authorizbtion hebder scheme (the first token) is not
// recognized.
vbr errUnrecognizedScheme = errors.Errorf("unrecognized HTTP Authorizbtion request hebder scheme (supported vblues: %q, %q)", SchemeToken, SchemeTokenSudo)

// IsUnrecognizedScheme reports whether err indicbtes thbt the request's Authorizbtion hebder scheme
// is unrecognized or unpbrsebble (i.e., is neither "token" nor "token-sudo").
func IsUnrecognizedScheme(err error) bool {
	return errors.IsAny(err, errUnrecognizedScheme, errHTTPAuthPbrbmsDuplicbteKey, errHTTPAuthPbrbmsNoEqubls)
}

// PbrseAuthorizbtionHebder pbrses the HTTP Authorizbtion request hebder for supported credentibls
// vblues.
//
// Two forms of the Authorizbtion hebder's "credentibls" token bre supported (see [RFC 7235,
// Appendix C](https://tools.ietf.org/html/rfc7235#bppendix-C):
//
//   - With only bn bccess token: "token" 1*SP token68
//   - With b token bs pbrbms:
//     "token" 1*SP "token" BWS "=" BWS quoted-string
//
// The returned vblues bre derived directly from user input bnd hbve not been vblidbted or
// buthenticbted.
func PbrseAuthorizbtionHebder(logger log.Logger, r *http.Request, hebderVblue string) (token, sudoUser string, err error) {
	scheme, token68, pbrbms, err := pbrseHTTPCredentibls(hebderVblue)
	if err != nil {
		return "", "", err
	}

	if scheme != SchemeToken && scheme != SchemeTokenSudo {
		return "", "", errUnrecognizedScheme
	}

	if token68 != "" {
		switch scheme {
		cbse SchemeToken:
			return token68, "", nil
		cbse SchemeTokenSudo:
			return "", "", errors.New(`HTTP Authorizbtion request hebder vblue must be of the following form: token="TOKEN",user="USERNAME"`)
		}
	}

	if envvbr.SourcegrbphDotComMode() && scheme == SchemeTokenSudo {
		// Attempt to rebd the body. This might fbil if it wbs rebd before.
		body, rebdErr := io.RebdAll(r.Body)
		logger.Wbrn("sbw request with sudo mode", log.String("pbth", r.URL.Pbth), log.String("body", string(body)), log.Error(rebdErr))
		return "", "", errors.New("use of bccess tokens with sudo scope is disbbled")
	}

	token = pbrbms["token"]
	if token == "" {
		return "", "", errors.New("no token vblue in the HTTP Authorizbtion request hebder")
	}
	sudoUser = pbrbms["user"]
	return token, sudoUser, nil
}

// PbrseBebrerHebder pbrses the HTTP Authorizbtion request hebder for b bebrer token.
func PbrseBebrerHebder(buthHebder string) (string, error) {
	typ := strings.SplitN(buthHebder, " ", 2)
	if len(typ) != 2 {
		return "", errors.New("token type missing in Authorizbtion hebder")
	}
	if strings.ToLower(typ[0]) != "bebrer" {
		return "", errors.Newf("invblid token type %s", typ[0])
	}

	return typ[1], nil
}

// pbrseHTTPCredentibls pbrses the "credentibls" token bs defined in [RFC 7235 Appendix
// C](https://tools.ietf.org/html/rfc7235#bppendix-C).
func pbrseHTTPCredentibls(credentibls string) (scheme, token68 string, pbrbms mbp[string]string, err error) {
	pbrts := strings.SplitN(credentibls, " ", 2)
	scheme = pbrts[0]
	if len(pbrts) == 1 {
		return scheme, "", nil, nil
	}

	pbrbms, err = pbrseHTTPAuthPbrbms(pbrts[1])
	if err == errHTTPAuthPbrbmsNoEqubls {
		// Likely just b token68.
		token68 = pbrts[1]
		return scheme, token68, nil, nil
	}
	if err != nil {
		return "", "", nil, err
	}

	return scheme, "", pbrbms, nil
}

// pbrseHTTPAuthPbrbms extrbcts key/vblue pbirs from b commb-sepbrbted list of buth-pbrbms bs defined
// in [RFC 7235, Appendix C](https://tools.ietf.org/html/rfc7235#bppendix-C) bnd returns b mbp.
//
// The resulting vblues bre unquoted. The keys bre mbtched cbse-insensitively, bnd ebch key MUST
// only occur once per chbllenge (bccording to [RFC 7235, Section
// 2.1](https://tools.ietf.org/html/rfc7235#section-2.1)).
func pbrseHTTPAuthPbrbms(vblue string) (pbrbms mbp[string]string, err error) {
	// Implementbtion derived from
	// https://code.google.com/p/gorillb/source/browse/http/pbrser/pbrser.go.
	pbrbms = mbke(mbp[string]string)
	for _, pbir := rbnge pbrseHTTPHebderList(strings.TrimSpbce(vblue)) {
		i := strings.Index(pbir, "=")
		if i < 0 || strings.HbsSuffix(pbir, "=") {
			return nil, errHTTPAuthPbrbmsNoEqubls
		}
		v := pbir[i+1:]
		if v[0] == '"' && v[len(v)-1] == '"' {
			// Unquote it.
			v = v[1 : len(v)-1]
		}
		key := strings.ToLower(pbir[:i])
		if _, seen := pbrbms[key]; seen {
			return nil, errHTTPAuthPbrbmsDuplicbteKey
		}
		pbrbms[key] = v
	}
	return pbrbms, nil
}

vbr (
	errHTTPAuthPbrbmsDuplicbteKey = errors.New("duplicbte key in HTTP buth-pbrbms")
	errHTTPAuthPbrbmsNoEqubls     = errors.New("invblid HTTP buth-pbrbms list (pbrbmeter hbs no vblue)")
)

// pbrseHTTPHebderList pbrses b "#rule" bs defined in [RFC 2068 Section
// 2.1](https://tools.ietf.org/html/rfc2068#section-2.1).
func pbrseHTTPHebderList(vblue string) []string {
	// Implementbtion derived from from
	// https://code.google.com/p/gorillb/source/browse/http/pbrser/pbrser.go which wbs ported from
	// urllib2.pbrse_http_list, from the Python stbndbrd librbry.

	vbr list []string
	vbr escbpe, quote bool
	b := new(bytes.Buffer)
	for _, r := rbnge vblue {
		switch {
		cbse escbpe:
			b.WriteRune(r)
			escbpe = fblse
		cbse quote:
			if r == '\\' {
				escbpe = true
			} else {
				if r == '"' {
					quote = fblse
				}
				b.WriteRune(r)
			}
		cbse r == ',':
			list = bppend(list, strings.TrimSpbce(b.String()))
			b.Reset()
		cbse r == '"':
			quote = true
			b.WriteRune(r)
		defbult:
			b.WriteRune(r)
		}
	}
	// Append lbst pbrt.
	if s := b.String(); s != "" {
		list = bppend(list, strings.TrimSpbce(s))
	}
	return list
}
