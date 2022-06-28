package check

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/sync/semaphore"
)

type RunFunc func(ctx context.Context) (string, error)

type Check struct {
	Name        string
	Description string

	Run RunFunc

	LastRun           time.Time
	CachedCheckErr    error
	CachedCheckOutput string
}

func (c *Check) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name              string    `json:"name"`
		Description       string    `json:"description"`
		LastRun           time.Time `json:"last_run"`
		CachedCheckErr    string    `json:"cached_check_err"`
		CachedCheckOutput string    `json:"cached_check_output"`
	}{
		Name:              c.Name,
		Description:       c.Description,
		LastRun:           c.LastRun,
		CachedCheckErr:    errString(c.CachedCheckErr),
		CachedCheckOutput: c.CachedCheckOutput,
	})
}

type HealthChecker struct {
	mu     sync.Mutex
	Checks []Check
}

func (c *HealthChecker) MarshalJSON() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return json.Marshal(c.Checks)
}

func (c *HealthChecker) Run() {
	for {
		time.Sleep(60 * time.Second)
		c.doRun()
	}
}

func (c *HealthChecker) doRun() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	sem := semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
	wg := sync.WaitGroup{}

	for i, check := range c.Checks {
		err := sem.Acquire(ctx, 1)
		if err != nil {
			return
		}
		wg.Add(1)
		go func(i int, check Check) {
			defer sem.Release(1)
			defer wg.Done()

			check.CachedCheckOutput, check.CachedCheckErr = check.Run(ctx)
			check.LastRun = time.Now().UTC()

			c.mu.Lock()
			c.Checks[i] = check
			c.mu.Unlock()
		}(i, check)
	}
	wg.Wait()
}

func (c *HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case "application/json":
		out, err := json.Marshal(c)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(out)))

		w.Write(out)

	default:
		out, err := c.RenderTable()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to render plain text table: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(out)))

		w.Write(out)
	}
}

func (c *HealthChecker) RenderTable() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	bw := bytes.Buffer{}
	tw := tabwriter.NewWriter(&bw, 16, 8, 4, ' ', 0)

	_, err := fmt.Fprintf(tw, "status\tname\tdescription\toutput\terr\tlast_run\n")
	if err != nil {
		return nil, errors.Wrap(err, "writing column headers")
	}

	for _, check := range c.Checks {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			func() string {
				if check.LastRun.IsZero() {
					return "PENDING"
				}
				if check.CachedCheckErr == nil {
					return "OK"
				}
				return "FAIL"
			}(),
			check.Name,
			check.Description,
			check.CachedCheckOutput,
			errString(check.CachedCheckErr),
			timestampString(check.LastRun),
		)
	}

	err = tw.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "flushing tabwriter")
	}

	return bw.Bytes(), nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func timestampString(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.RFC3339)
}
