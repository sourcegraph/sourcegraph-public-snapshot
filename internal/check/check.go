package check

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// RunFunc returns a valid JSON and an error. The HealthChecker will call
// json.Marshal on any.
type RunFunc func(ctx context.Context) (any, error)

type Check struct {
	// Name should have the prefix "check_".
	Name string

	// How often the check runs.
	Interval time.Duration

	Run RunFunc
}

type HealthChecker struct {
	Checks []Check
}

const timeout = time.Second * 60

func (hc *HealthChecker) Init() {
	for _, check := range hc.Checks {
		go func(c Check) {

			// Each check is represented by an expvar.Map with the fields
			// - status OK|FAIL|PENDING
			// - out JSON
			// - error STRING
			// - last_run RFC3339 formatted timestamp
			m := expvar.NewMap(c.Name)

			status := new(expvar.String)
			status.Set("PENDING")
			m.Set("status", status)

			out := new(expvar.String)
			m.Set("out", out)

			checkErr := new(expvar.String)
			m.Set("error", checkErr)

			lastRun := new(expvar.String)
			// set to 0 to make it easier for the client to parse.
			lastRun.Set(time.Time{}.Format(time.RFC3339))
			m.Set("last_run", lastRun)

			for {
				time.Sleep(c.Interval)

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				res, err := c.Run(ctx)
				if err != nil {
					status.Set("FAIL")
				} else {
					status.Set("OK")
				}

				lastRun.Set(time.Now().Format(time.RFC3339))

				checkErr.Set(errString(err))

				b, _ := json.Marshal(res)
				out.Set(string(b))
			}
		}(check)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type NewHealthChecksHandler func() http.Handler

// NewHealthCheckHandler returns a handler that serves a page just like /vars
// but filtered for health checks. Each service should expose a /checks
// endpoint. All /checks endpoints are aggregates by ServeHealthCheckAggregate
func (hc *HealthChecker) NewHealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		first := true
		for _, check := range hc.Checks {
			if !first {
				fmt.Fprintf(w, ",\n")
			}
			first = false

			v := expvar.Get(check.Name)
			fmt.Fprintf(w, "%q: %s", check.Name, v.String())
		}
		fmt.Fprintf(w, "\n}\n")
	})
}

// NewAggregateHealthCheckHandler should only be called in frontend.
func NewAggregateHealthCheckHandler() http.Handler {
	services := []string{"frontend"}

	endpoints := func(service string) []*url.URL {
		switch service {
		case "frontend":
			u, err := url.Parse(internalapi.Client.URL)
			if err != nil {
				return nil
			}
			return []*url.URL{u.ResolveReference(&url.URL{Path: "/.internal/checks"})}
		default:
			return nil
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		defer fmt.Fprintf(w, "\n}\n")

		firstService := true
		for _, service := range services {
			if !firstService {
				fmt.Fprintf(w, ",\n")
			}
			firstService = false

			fmt.Fprintf(w, "%q: {\n", service)

			firstEndpoint := true
			for _, endpoint := range endpoints(service) {
				if !firstEndpoint {
					fmt.Fprintf(w, ",\n")
				}
				firstEndpoint = false
				req, err := http.NewRequest("GET", endpoint.String(), nil)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res, err := httpcli.InternalDoer.Do(req)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer res.Body.Close()

				if res.StatusCode != 200 {
					fmt.Fprintf(w, "%q: %q\n", endpoint.Host, "unreachable")
					continue
				}

				b, err := io.ReadAll(res.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, "%q: %s\n", endpoint.Host, b)
			}

			fmt.Fprintf(w, "\n}\n")
		}
	})
}
