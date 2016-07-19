// Package depprof is an experiment in creating a dependency graph by profiling at runtime.
package depprof

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type handler struct {
	filterPrefix string

	depsMutex sync.Mutex
	deps      map[[2]string]struct{}
}

// NewHandler starts profiling all packages that match the given filterPrefix and returns an
// HTTP handler which gives access to the data collected.
func NewHandler(filterPrefix string) http.Handler {
	h := &handler{
		filterPrefix: strings.TrimRight(filterPrefix, "/") + "/",
		deps:         make(map[[2]string]struct{}),
	}
	go h.recordLoop()
	return h
}

func (h *handler) recordLoop() {
	var p []runtime.StackRecord
	var n int
	for {
		for {
			var ok bool
			n, ok = runtime.GoroutineProfile(p)
			if ok {
				break
			}
			p = make([]runtime.StackRecord, n+10)
		}

		h.depsMutex.Lock()
		for _, sr := range p[:n] {
			prevPkg := ""
			for _, pc := range sr.Stack() {
				file, _ := runtime.FuncForPC(pc).FileLine(pc)
				pkg, ok := fileToPkg(file)
				if !ok || !strings.HasPrefix(pkg, h.filterPrefix) {
					continue
				}
				if prevPkg != "" && pkg != prevPkg {
					h.deps[[2]string{pkg, prevPkg}] = struct{}{}
				}
				prevPkg = pkg
			}
		}
		h.depsMutex.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("show") {
	case "graph":
		h.depsMutex.Lock()
		defer h.depsMutex.Unlock()

		cmd := exec.Command("dot", "-Tsvg")
		cmd.Stdout = w
		cmd.Stderr = os.Stderr
		in, err := cmd.StdinPipe()
		if err != nil {
			panic(err)
		}

		go func() {
			defer in.Close()

			fmt.Fprintf(in, "digraph g {")
			for dep := range h.deps {
				fmt.Fprintf(in, `"%s" -> "%s";`, strings.TrimPrefix(dep[0], h.filterPrefix), strings.TrimPrefix(dep[1], h.filterPrefix))
			}
			fmt.Fprintf(in, "}")
		}()

		if err := cmd.Run(); err != nil {
			panic(err)
		}

	default:
		w.Write([]byte(`<a href="?show=graph">Graph</a>`))

	}
}

func fileToPkg(file string) (string, bool) {
	i := strings.Index(file, "/src/")
	if i == -1 {
		return "", false
	}
	j := strings.LastIndexByte(file, '/')
	return file[i+5 : j], true
}
