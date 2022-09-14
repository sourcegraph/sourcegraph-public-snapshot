package check

import (
	"bytes"
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

const (
	timeout = time.Second * 60

	ok      = "OK"
	fail    = "FAIL"
	pending = "PENDING"
)

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
			status.Set(pending)
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
					status.Set(fail)
				} else {
					status.Set(ok)
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

// ServeHTTP serves a page just like /vars but filtered for health checks. Each
// service should expose a /checks endpoint.
func (hc *HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
}

// TODO: How can we reach ALL frontends?
func DefaultEndpointProvider(service string) []*url.URL {
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

// NewAggregateHealthCheckHandler returns a JSON with the high-level structure
//
//	{
//	    <service1>: {
//	        <address1>: {
//	            <check1-name> : <check1-data>,
//	            <check2-name> : <check2-data>
//	        },
//	        <address2>: ...
//	    },
//	    <service2>: ...
//	    ...
//	}
//
// The handler should only be used in frontend.
//
// While it is not necessary to add newlines to the JSON output, it is really
// convenient for grepping from the command line.
func NewAggregateHealthCheckHandler(endpointProvider func(service string) []*url.URL) http.Handler {
	services := []string{"frontend"}

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

			endpoints := endpointProvider(service)
			if len(endpoints) == 0 {
				fmt.Fprintf(w, "%q: %q", service, fail+": no endpoints discovered")
				continue
			}

			fmt.Fprintf(w, "%q: {\n", service)

			firstEndpoint := true
			for _, endpoint := range endpoints {
				if !firstEndpoint {
					fmt.Fprintf(w, ",\n")
				}
				firstEndpoint = false

				// We can set aggressive timeouts, because the checks' results are cached on the
				// server.
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()

				req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				res, err := httpcli.InternalDoer.Do(req)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer res.Body.Close()

				if res.StatusCode != 200 {
					fmt.Fprintf(w, "%q: %q\n", endpoint.Host, fail+": unreachable")
					continue
				}

				b, err := io.ReadAll(res.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, "%q: %s", endpoint.Host, bytes.TrimSpace(b))
			}

			fmt.Fprintf(w, "\n}")
		}
	})
}
