package debugserver

import (
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"
)

// expvarHandler is copied from package expvar and exported so that it
// can be mounted on any ServeMux, not just http.DefaultServeMux.
func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_765(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
