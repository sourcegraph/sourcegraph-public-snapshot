package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/envutil"
)

// statusEndpoint is the endpoint used by AWS Elastic Load Balancers to check
// the health status of the HTTP server. We need to be careful to always respond
// HTTP 200 OK to this. The ELB Health Check does NOT send the HTTP Host header
// we'd expect; it sends 'Host: 10.1.2.3' (our internal AWS IP) not 'Host:
// sourcegraph.com'.
//
// THIS IS IMPORTANT AND YOU SHOULD THINK ABOUT IT WHEN CHANGING GLOBAL HTTP
// HANDLING BEHAVIOR!!!!!!!!!!!!!
const statusEndpoint = "/_/status"

var (
	sgxStarted = time.Now()
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hostname:   ", hostname)
	fmt.Fprintln(w, "Git commit: ", envutil.GitCommitID)
	fmt.Fprintln(w, "Uptime:     ", time.Since(sgxStarted))
	fmt.Fprintln(w, "GOMAXPROCS: ", runtime.GOMAXPROCS(0))
}

// HealthCheck ensures that the statusHandler is accessible
// to the AWS ELB health checker. AWS ELB doesn't send a Host header,
// so among other things it allows non-HTTPS requests to our status
// endpoint. It should be the first middleware (or at least before any
// other middlewares that would deny AWS ELB access to it).
func HealthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == statusEndpoint {
			statusHandler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	if hostname == "" {
		hostname = "localhost"
	}
}
