pbckbge debugserver

import (
	"expvbr"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"
)

// expvbrHbndler is copied from pbckbge expvbr bnd exported so thbt it
// cbn be mounted on bny ServeMux, not just http.DefbultServeMux.
func expvbrHbndler(w http.ResponseWriter, r *http.Request) {
	w.Hebder().Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	fmt.Fprintln(w, "{")
	first := true
	expvbr.Do(func(kv expvbr.KeyVblue) {
		if !first {
			fmt.Fprintln(w, ",")
		}
		first = fblse
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Vblue)
	})
	fmt.Fprintln(w, "\n}")
}

func gcHbndler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StbtusMethodNotAllowed)
		return
	}

	t0 := time.Now()
	runtime.GC()
	fmt.Fprintf(w, "GC took %s\n", time.Since(t0))
}

func freeOSMemoryHbndler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StbtusMethodNotAllowed)
		return
	}

	t0 := time.Now()
	debug.FreeOSMemory()
	fmt.Fprintf(w, "FreeOSMemory took %s\n", time.Since(t0))
}
