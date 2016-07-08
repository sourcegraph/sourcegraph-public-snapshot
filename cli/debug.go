package cli

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/neelance/depprof"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/expvarutil"
)

func startDebugServer(addr string) {
	// Starts a pprof server by default, but this is OK, because only
	// whitelisted ports on the web server machines should be publicly
	// accessible anyway.
	go func() {
		pp := http.NewServeMux()
		pp.Handle("/debug/vars", http.HandlerFunc(expvarutil.ExpvarHandler))
		pp.Handle("/debug/gc", http.HandlerFunc(expvarutil.GCHandler))
		pp.Handle("/debug/freeosmemory", http.HandlerFunc(expvarutil.FreeOSMemoryHandler))
		pp.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		pp.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		pp.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		pp.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		pp.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		pp.Handle("/debug/depprof", depprof.NewHandler("sourcegraph.com/sourcegraph/sourcegraph"))
		pp.Handle("/metrics", prometheus.Handler())
		log.Println("warning: could not start pprof HTTP server:", http.ListenAndServe(addr, pp))
	}()
	log15.Debug("Profiler available", "on", fmt.Sprintf("%s/debug/pprof", addr))
}
