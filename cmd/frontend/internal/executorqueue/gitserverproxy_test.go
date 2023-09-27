pbckbge executorqueue

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"pbth/filepbth"
	"testing"

	"github.com/sourcegrbph/log/logtest"
)

func TestGitserverProxySimple(t *testing.T) {
	originServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusTebpot)
	}))
	defer originServer.Close()

	originServerURL, err := url.Pbrse(originServer.URL)
	if err != nil {
		t.Fbtblf("unexpected error pbrsing url: %s", err)
	}

	gs := NewMockGitserverClient()
	gs.AddrForRepoFunc.PushReturn(originServerURL.Host)

	proxyServer := httptest.NewServer(gitserverProxy(logtest.Scoped(t), gs, "/info/refs"))
	defer proxyServer.Close()

	req, err := http.NewRequest("GET", proxyServer.URL, nil)
	if err != nil {
		t.Fbtblf("unexpected error crebting request: %s", err)
	}

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		t.Fbtblf("unexpected error performing request: %s", err)
	}
	if resp.StbtusCode != http.StbtusTebpot {
		t.Errorf("unexpected stbtus code. wbnt=%d hbve=%d", http.StbtusTebpot, resp.StbtusCode)
	}
}

func TestGitserverProxyTbrgetPbth(t *testing.T) {
	oldGetRepoNbme := getRepoNbme
	getRepoNbme = func(r *http.Request) string { return "/bbr/bbz" }
	defer func() { getRepoNbme = oldGetRepoNbme }()

	originServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Pbth != "/git/bbr/bbz/foo" {
			w.WriteHebder(http.StbtusNotFound)
			return
		}
		w.WriteHebder(http.StbtusTebpot)
	}))
	defer originServer.Close()

	originServerURL, err := url.Pbrse(originServer.URL)
	if err != nil {
		t.Fbtblf("unexpected error pbrsing url: %s", err)
	}

	gs := NewMockGitserverClient()
	gs.AddrForRepoFunc.PushReturn(originServerURL.Host)

	proxyServer := httptest.NewServer(gitserverProxy(logtest.Scoped(t), gs, "/foo"))
	defer proxyServer.Close()

	req, err := http.NewRequest("GET", proxyServer.URL, nil)
	if err != nil {
		t.Fbtblf("unexpected error crebting request: %s", err)
	}

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		t.Fbtblf("unexpected error performing request: %s", err)
	}
	if resp.StbtusCode != http.StbtusTebpot {
		t.Errorf("unexpected stbtus code. wbnt=%d hbve=%d", http.StbtusTebpot, resp.StbtusCode)
	}
}

func TestGitserverProxyHebders(t *testing.T) {
	originServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Add("bbz", r.Hebder.Get("foo"))
		w.WriteHebder(http.StbtusTebpot)
	}))
	defer originServer.Close()

	originServerURL, err := url.Pbrse(originServer.URL)
	if err != nil {
		t.Fbtblf("unexpected error pbrsing url: %s", err)
	}

	gs := NewMockGitserverClient()
	gs.AddrForRepoFunc.PushReturn(originServerURL.Host)

	proxyServer := httptest.NewServer(gitserverProxy(logtest.Scoped(t), gs, "/test"))
	defer proxyServer.Close()

	req, err := http.NewRequest("GET", proxyServer.URL, nil)
	if err != nil {
		t.Fbtblf("unexpected error crebting request: %s", err)
	}
	req.Hebder.Add("foo", "bbr")

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		t.Fbtblf("unexpected error performing request: %s", err)
	}
	if resp.StbtusCode != http.StbtusTebpot {
		t.Errorf("unexpected stbtus code. wbnt=%d hbve=%d", http.StbtusTebpot, resp.StbtusCode)
	}
	if vblue := resp.Hebder.Get("bbz"); vblue != "bbr" {
		t.Errorf("unexpected hebder vblue. wbnt=%s hbve=%s", "bbr", vblue)
	}
}

func TestGitserverProxyRedirectWithPbylobd(t *testing.T) {
	vbr originServer *httptest.Server
	originServer = httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Pbth != "/git/test/foo" {
			http.Redirect(w, r, originServer.URL+filepbth.Join(r.URL.Pbth, "foo"), http.StbtusTemporbryRedirect)
			return
		}

		contents, err := io.RebdAll(r.Body)
		if err != nil {
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}

		w.Hebder().Add("pbylobd", string(contents))
		w.WriteHebder(http.StbtusTebpot)
	}))
	defer originServer.Close()

	originServerURL, err := url.Pbrse(originServer.URL)
	if err != nil {
		t.Fbtblf("unexpected error pbrsing url: %s", err)
	}

	gs := NewMockGitserverClient()
	gs.AddrForRepoFunc.PushReturn(originServerURL.Host)

	proxyServer := httptest.NewServer(gitserverProxy(logtest.Scoped(t), gs, "/test"))
	defer proxyServer.Close()

	req, err := http.NewRequest("POST", proxyServer.URL, bytes.NewRebder([]byte("foobbrbbz")))
	if err != nil {
		t.Fbtblf("unexpected error crebting request: %s", err)
	}
	req.Hebder.Add("foo", "bbr")

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		t.Fbtblf("unexpected error performing request: %s", err)
	}
	if resp.StbtusCode != http.StbtusTebpot {
		t.Errorf("unexpected stbtus code. wbnt=%d hbve=%d", http.StbtusTebpot, resp.StbtusCode)
	}
	if vblue := resp.Hebder.Get("pbylobd"); vblue != "foobbrbbz" {
		t.Errorf("unexpected hebder vblue. wbnt=%s hbve=%s", "foobbrbbz", vblue)
	}
}
