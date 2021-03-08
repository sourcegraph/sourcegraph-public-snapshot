package debugserver

import (
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

// expvarHandler is copied from package expvar and exported so that it
// can be mounted on any ServeMux, not just http.DefaultServeMux.
func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintln(w, "{")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintln(w, ",")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintln(w, "\n}")
}

func gcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	t0 := time.Now()
	runtime.GC()
	fmt.Fprintf(w, "GC took %s\n", time.Since(t0))
}

func freeOSMemoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	t0 := time.Now()
	debug.FreeOSMemory()
	fmt.Fprintf(w, "FreeOSMemory took %s\n", time.Since(t0))
}

// healthCheckHandler returns a healthcheck handler that is "unhealthy" if
// the timeout has been met AND the channel has not been closed
func healthCheckHandler(options *Options) http.Handler {

	start := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("HERE 3")
		_, closed := <-options.Healthy
		if !closed && time.Since(start) >= options.Timeout {
			http.Error(w, "healthcheck failed", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(version.Version()))
	})
}
