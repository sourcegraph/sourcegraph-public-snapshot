package load

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/e2etest/e2etestuser"

	vegeta "github.com/tsenart/vegeta/lib"

	"golang.org/x/net/context"
)

// LoadTest can Run an individual load test
type LoadTest struct {
	// Endpoint is Who we are testing (eg "sourcegraph.com")
	Endpoint *url.URL

	AttackerOpts []func(*vegeta.Attacker)
	TargetPaths  []string
	Rate         uint64
	ReportPeriod time.Duration

	Username  string
	Password  string
	Anonymous bool
}

// Run runs the load test until the auth token expires
func (t *LoadTest) Run(ctx context.Context) error {
	var testDuration time.Duration
	hdr := http.Header{}
	hdr.Set("User-Agent", e2etestuser.UserAgent)

	cookie, err := t.getCookie()
	if err != nil {
		return err
	} else if cookie != nil {
		hdr.Set("Cookie", cookie.HeaderValue)
		testDuration = cookie.Expires
	}

	atk := vegeta.NewAttacker(t.AttackerOpts...)
	tr, err := t.targeter(hdr)
	if err != nil {
		return err
	}

	log.Printf("Starting %v", t)
	res := atk.Attack(tr, t.Rate, testDuration)
	defer atk.Stop()

	mAll := &vegeta.Metrics{}
	mPartial := &vegeta.Metrics{}
	reportTicker := time.NewTicker(t.ReportPeriod)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping %v", t)
		case <-reportTicker.C:
			log.Printf("Report for the last %s:", t.ReportPeriod)
			t.report(mPartial)
			mPartial = &vegeta.Metrics{}
			continue
		case r, ok := <-res:
			if ok {
				mAll.Add(r)
				mPartial.Add(r)
				continue
			}
		}
		break
	}

	log.Printf("Finished %v", t)
	log.Printf("Report for %s:", t)
	t.report(mAll)
	return ctx.Err()
}

func (t *LoadTest) getCookie() (*authedCookie, error) {
	if t.Anonymous {
		return nil, nil
	}
	if t.Username == "" {
		err := createLoadTestUser(t.Endpoint)
		if err != nil {
			return nil, err
		}
		t.Username = testUserName
		t.Password = testPassword
	}
	return getAuthedCookie(t.Endpoint, t.Username, t.Password)
}

func (t *LoadTest) targeter(hdr http.Header) (vegeta.Targeter, error) {
	targets := make([]WeightedTarget, len(t.TargetPaths))
	for i, p := range t.TargetPaths {
		f := strings.Fields(p)
		if len(f) == 1 {
			f = append(f, "1")
		}
		url := t.Endpoint.String() + f[0]
		weight, err := strconv.ParseFloat(f[1], 64)
		if err != nil {
			return nil, fmt.Errorf("Weight %q is not a float. Target %q", f[1], p)
		}

		targets[i] = WeightedTarget{
			Target: vegeta.Target{
				Method: "GET",
				URL:    url,
				Header: hdr,
			},
			Weight: weight,
		}
		log.Printf("Target: %gx %s", weight, url)
	}
	return NewWeightedTargeter(targets...), nil
}

func (t *LoadTest) report(m *vegeta.Metrics) {
	m.Close()
	m.Errors = []string{} // Ignore noisy error list
	vegeta.NewTextReporter(m)(os.Stderr)
}

func (t *LoadTest) String() string {
	return fmt.Sprintf("LoadTest{Endpoint=%v, Username=%v, Rate=%v}", t.Endpoint, t.Username, t.Rate)
}
