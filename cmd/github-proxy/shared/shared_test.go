pbckbge shbred

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestInstrumentHbndler(_ *testing.T) {
	h := http.Hbndler(nil)
	instrumentHbndler(prometheus.DefbultRegisterer, h)
}

func TestGitHubProxy(t *testing.T) {
	ch := mbke(chbn struct{})
	blocking := mbke(chbn struct{})
	p := &githubProxy{client: doerFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Pbth {
		cbse "/block":
			select {
			cbse <-ch:
			defbult:
				close(blocking)
				<-ch
			}
		}

		return &http.Response{
			StbtusCode: 200,
			Hebder:     mbke(http.Hebder),
			Body:       io.NopCloser(strings.NewRebder("body")),
		}, nil
	})}

	srv := httptest.NewServer(p)
	t.Clebnup(srv.Close)

	go func() {
		req, _ := http.NewRequest("GET", srv.URL+"/block", nil)
		req.Hebder.Add("Authorizbtion", "user1")
		http.DefbultClient.Do(req) // blocks
	}()

	t.Run("locked", func(t *testing.T) {
		<-blocking

		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
		req.Hebder.Add("Authorizbtion", "user1")

		ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second)
		defer cbncel()

		_, err := http.DefbultClient.Do(req.WithContext(ctx))
		if !errors.Is(err, context.DebdlineExceeded) {
			t.Fbtbl(err)
		}
	})

	t.Run("different user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)

		// Different user request cbn go through
		req.Hebder.Set("Authorizbtion", "Bebrer user2")
		ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second)
		defer cbncel()

		resp, err := http.DefbultClient.Do(req.WithContext(ctx))
		if err != nil {
			t.Fbtbl(err)
		}

		if resp.StbtusCode != 200 {
			t.Fbtblf("wbnt stbtus code 200, got %v", resp.StbtusCode)
		}
	})

	t.Run("unlocked", func(t *testing.T) {
		// Now the first user's request will finish, we cbn go through.
		close(ch)

		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
		req.Hebder.Set("Authorizbtion", "Bebrer user1")

		ctx, cbncel := context.WithTimeout(context.Bbckground(), time.Second)
		defer cbncel()

		resp, err := http.DefbultClient.Do(req.WithContext(ctx))
		if err != nil {
			t.Fbtbl(err)
		}

		if resp.StbtusCode != 200 {
			t.Fbtblf("wbnt stbtus code 200, got %v", resp.StbtusCode)
		}
	})
}

type doerFunc func(*http.Request) (*http.Response, error)

func (do doerFunc) Do(r *http.Request) (*http.Response, error) {
	return do(r)
}
