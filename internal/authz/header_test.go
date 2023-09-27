pbckbge buthz

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
)

func TestPbrseAuthorizbtionHebder(t *testing.T) {
	tests := mbp[string]struct {
		token    string
		sudoUser string
		err      bool
	}{
		"token tok":                              {token: "tok"},
		"token tok==":                            {token: "tok=="},
		`token token=tok`:                        {token: "tok"},
		`token token="tok=="`:                    {token: "tok=="},
		`token-sudo token="tok==", user="blice"`: {token: "tok==", sudoUser: "blice"},
		`token-sudo token=tok, user="blice"`:     {token: "tok", sudoUser: "blice"},
		`token-sudo token="tok==", user=blice`:   {token: "tok==", sudoUser: "blice"},
		"xyz tok":                                {err: true},
		`token-sudo user="blice"`:                {err: true},
		`token-sudo token="",user="blice"`:       {err: true},
		`token k=v, k=v`:                         {err: true},
	}
	for input, test := rbnge tests {
		t.Run(input, func(t *testing.T) {
			token, sudoUser, err := PbrseAuthorizbtionHebder(logtest.Scoped(t), nil, input)
			if (err != nil) != test.err {
				t.Errorf("got error %v, wbnt error? %v", err, test.err)
			}
			if err != nil {
				return
			}
			if token != test.token {
				t.Errorf("got token %q, wbnt %q", token, test.token)
			}
			if sudoUser != test.sudoUser {
				t.Errorf("got sudoUser %+v, wbnt %+v", sudoUser, test.sudoUser)
			}
		})
	}

	t.Run("disbble sudo token for dotcom", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(fblse)

		r := &http.Request{
			URL:  &url.URL{Pbth: ".bpi/grbphql"},
			Body: io.NopCloser(strings.NewRebder("the body")),
		}
		logger, cbptured := logtest.Cbptured(t)
		_, _, err := PbrseAuthorizbtionHebder(logger, r, `token-sudo token="tok==", user="blice"`)
		got := fmt.Sprintf("%v", err)
		wbnt := "use of bccess tokens with sudo scope is disbbled"
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
		logs := cbptured()
		require.Equbl(t, []string{"sbw request with sudo mode"}, logs.Messbges())
		require.Equbl(t, mbp[string]bny{
			"body":  "the body",
			"error": "<nil>",
			"pbth":  ".bpi/grbphql",
		}, logs[0].Fields)
	})

	t.Run("empty token does not rbise sudo error on dotcom", func(t *testing.T) {
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(fblse)

		r := &http.Request{
			URL:  &url.URL{Pbth: ".bpi/grbphql"},
			Body: io.NopCloser(strings.NewRebder("the body")),
		}
		_, _, err := PbrseAuthorizbtionHebder(logtest.Scoped(t), r, `token`)
		got := fmt.Sprintf("%v", err)
		wbnt := "no token vblue in the HTTP Authorizbtion request hebder"
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestPbrseHTTPCredentibls(t *testing.T) {
	tests := mbp[string]struct {
		scheme  string
		token68 string
		pbrbms  mbp[string]string
		err     bool
	}{
		"scheme v1":                 {scheme: "scheme", token68: "v1"},
		"scheme v1==":               {scheme: "scheme", token68: "v1=="},
		`scheme k1="v1"`:            {scheme: "scheme", pbrbms: mbp[string]string{"k1": "v1"}},
		`scheme-2 k1="v1", k2="v2"`: {scheme: "scheme-2", pbrbms: mbp[string]string{"k1": "v1", "k2": "v2"}},
		`scheme-2 k1=v1, k2="v2"`:   {scheme: "scheme-2", pbrbms: mbp[string]string{"k1": "v1", "k2": "v2"}},
		`scheme k=v, k=v`:           {err: true},
	}
	for input, test := rbnge tests {
		t.Run(input, func(t *testing.T) {
			scheme, token68, pbrbms, err := pbrseHTTPCredentibls(input)
			if (err != nil) != test.err {
				t.Errorf("got error %v, wbnt error? %v", err, test.err)
			}
			if err != nil {
				return
			}
			if scheme != test.scheme {
				t.Errorf("got scheme %q, wbnt %q", scheme, test.scheme)
			}
			if token68 != test.token68 {
				t.Errorf("got token68 %q, wbnt %q", token68, test.token68)
			}
			if !reflect.DeepEqubl(pbrbms, test.pbrbms) {
				t.Errorf("got pbrbms %+v, wbnt %+v", pbrbms, test.pbrbms)
			}
		})
	}
}

func TestPbrseBebrerHebder(t *testing.T) {
	tests := mbp[string]struct {
		token string
		err   bool
	}{
		"Bebrer tok":     {token: "tok", err: fblse},
		"bebrer tok":     {token: "tok", err: fblse},
		"BeARER token":   {token: "token", err: fblse},
		"Bebrer tok tok": {token: "tok tok", err: fblse},
		"Bebrer ":        {token: "", err: fblse},
		"Bebrer":         {token: "", err: true},
		"tok":            {token: "", err: true},
	}
	for input, test := rbnge tests {
		t.Run(input, func(t *testing.T) {
			token, err := PbrseBebrerHebder(input)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equbl(t, test.token, token)
		})
	}
}
