package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"context"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/graphqlbackend"

	"github.com/prometheus/client_golang/prometheus"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
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
	if _, err := graphqlbackend.ResolveRepo(r.Context(), data.Op.New.URI); err != nil {
		return err
	}
	w.Write([]byte("OK"))
	return nil
}

func resolveLocalRepo(ctx context.Context, repoPath string) (int32, error) {
	return handlerutil.GetRepoID(ctx, map[string]string{"Repo": repoPath})
}

func resolveLocalRepoRev(ctx context.Context, repoRev routevar.RepoRev) (*sourcegraph.RepoRevSpec, error) {
	repo, err := resolveLocalRepo(ctx, repoRev.Repo)
	if err != nil {
		return nil, err
	}
	res, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo, Rev: repoRev.Rev})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.RepoRevSpec{
		Repo:     repo,
		CommitID: res.CommitID,
	}, nil
}

func resolveLocalRepos(ctx context.Context, repoPaths []string, ignoreErrors bool) ([]int32, error) {
	repoIDs := make([]int32, 0, len(repoPaths))
	for _, repoPath := range repoPaths {
		repoID, err := resolveLocalRepo(ctx, repoPath)
		if err != nil {
			if !ignoreErrors {
				return nil, err
			} else {
				log15.Warn("resolve local repo", "err", err, "repo", repoPath)
			}
		} else {
			repoIDs = append(repoIDs, repoID)
		}
	}
	return repoIDs, nil
}

// repoIDOrPath is a type used purely for documentation purposes to
// indicate that this URL query parameter can be either a string (repo
// path) or number (repo ID).
type repoIDOrPath string

func init() {
	schemaDecoder.RegisterConverter(repoIDOrPath(""), func(s string) reflect.Value { return reflect.ValueOf(s) })
}

// getRepoID gets the repo ID from an interface{} type that can be
// either an int32 or a string. Typically callers decode the URL query
// string's "Repo" or "repo" field into an interface{} value and then
// pass it to getRepoID. This way, they can accept either the numeric
// repo ID or the repo path, which presents a nicer API to consumers.
func getRepoID(ctx context.Context, v repoIDOrPath) (int32, error) {
	if n, err := strconv.Atoi(string(v)); err == nil {
		return int32(n), nil
	}
	return handlerutil.GetRepoID(ctx, map[string]string{"Repo": string(v)})
}
