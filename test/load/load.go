package load

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	vegeta "github.com/tsenart/vegeta/lib"

	"golang.org/x/net/context"
)

type LoadTest struct {
	// Endpoint is Who we are testing (eg "sourcegraph.com")
	Endpoint *url.URL

	AttackerOpts []func(*vegeta.Attacker)
	TargetPaths  []string
	Rate         uint64

	Username string
	Password string
}

// Run runs the load test until the auth token expires
func (t *LoadTest) Run(ctx context.Context) error {
	// We need the cookie to do load tests as an authenticated user
	cookie, err := getAuthedCookie(t.Endpoint, t.Username, t.Password)
	if err != nil {
		return err
	}
	hdr := http.Header{}
	hdr.Set("Cookie", cookie.HeaderValue)
	hdr.Set("User-Agent", "Sourcegraph-Load-Test/0.1")

	atk := vegeta.NewAttacker(t.AttackerOpts...)
	tr := t.targeter(hdr)

	log.Printf("Starting %v", t)
	res := atk.Attack(tr, t.Rate, cookie.Expires)
	defer atk.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping %v", t)
		case r, ok := <-res:
			if ok {
				log.Println(*r)
				continue
			}
		}
		break
	}

	log.Printf("Finished %v", t)
	// TODO dump some stats
	return ctx.Err()
}

func (t *LoadTest) targeter(hdr http.Header) vegeta.Targeter {
	targets := make([]vegeta.Target, len(t.TargetPaths))
	for i, p := range t.TargetPaths {
		url := *t.Endpoint
		url.Path = p
		targets[i] = vegeta.Target{
			Method: "GET",
			URL:    url.String(),
			Header: hdr,
		}
		log.Println("Target:", targets[i].URL)
	}
	return vegeta.NewStaticTargeter(targets...)
}

func (t *LoadTest) String() string {
	return fmt.Sprintf("LoadTest{Endpoint=%v, Username=%v, Rate=%v}", t.Endpoint, t.Username, t.Rate)
}
