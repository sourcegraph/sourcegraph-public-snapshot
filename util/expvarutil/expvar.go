package expvarutil

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// ExpvarHandler is copied from package expvar and exported so that it
// can be mounted on any ServeMux, not just http.DefaultServeMux.
func ExpvarHandler(w http.ResponseWriter, r *http.Request) {
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

func GCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	t0 := time.Now()
	runtime.GC()
	fmt.Fprintf(w, "GC took %s\n", time.Since(t0))
}

// NewClient returns a new remote expvar client; varsURL is the URL to
// the ExpvarHandler on a remote web server, such as
// http://localhost:6060/debug/vars.
func NewClient(varsURL string) *Client {
	return &Client{varsURL}
}

type Client struct {
	varsURL string
}

// GC causes the remote server to GC. This is useful since many
// expvars are related to memory allocation, and GCing prior to
// getting memory stats reduces noise in the data.
//
// This assumes the server has the GCHandler mounted as a sibling to
// /debug/vars at /debug/gc.
func (c *Client) GC() error {
	url := strings.TrimSuffix(c.varsURL, "/vars") + "/gc"
	resp, err := http.Post(url, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error status %d (URL: %s)", resp.StatusCode, resp.Request.URL)
	}
	return nil
}

// Get gets a remote expvar's value and JSON-decodes it into v.
func (c *Client) Get(name string, v interface{}) error {
	resp, err := http.Get(c.varsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return err
	}

	vmap := map[string]json.RawMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&vmap); err != nil {
		return err
	}

	return json.Unmarshal(vmap[name], v)
}

func FreeOSMemoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	t0 := time.Now()
	debug.FreeOSMemory()
	fmt.Fprintf(w, "FreeOSMemory took %s\n", time.Since(t0))
}
