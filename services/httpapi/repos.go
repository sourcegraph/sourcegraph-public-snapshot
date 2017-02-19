package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"

	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var repoSearchDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "repos_list",
	Name:      "search_duration_seconds",
	Help:      "Repo search latency in seconds.",
	// Buckets are similar to statsutil.UserLatencyBuckets, but with more granularity for apdex measurements.
	Buckets: []float64{0.1, 0.2, 0.5, 0.8, 1, 1.5, 2, 5, 10, 15, 20, 30},
}, []string{"success", "query", "remote_search", "remote_only"})

func init() {
	prometheus.MustRegister(repoSearchDuration)
}

func serveRepos(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	start := time.Now()
	defer func() {
		duration := time.Now().Sub(start)
		labels := prometheus.Labels{
			"success":       fmt.Sprintf("%t", err == nil),
			"query":         fmt.Sprint(opt.Query != ""),
			"remote_search": fmt.Sprint(opt.RemoteSearch),
			"remote_only":   fmt.Sprint(opt.RemoteOnly),
		}
		repoSearchDuration.With(labels).Observe(duration.Seconds())
	}()

	repos, err := backend.Repos.List(r.Context(), &opt)
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, 60*time.Second); clientCached || err != nil {
		return err
	}

	return writeJSON(w, repos)
}

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	// legacy support for Chrome extension
	var data struct {
		Op struct {
			New struct {
				URI string
			}
		}
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return err
	}
	if _, err := backend.Repos.GetByURI(r.Context(), data.Op.New.URI); err != nil {
		return err
	}
	w.Write([]byte("OK"))
	return nil
}
