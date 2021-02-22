package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func serveGraphQL(schema *graphql.Schema, isInternal bool) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		if r.Method != "POST" {
			// The URL router should not have routed to this handler if method is not POST, but just in
			// case.
			return errors.New("method must be POST")
		}

		// We use the query to denote the name of a GraphQL request, e.g. for /.api/graphql?Repositories
		// the name is "Repositories".
		requestName := "unknown"
		if r.URL.RawQuery != "" {
			requestName = r.URL.RawQuery
		}
		requestSource := guessSource(r)

		// Used by the prometheus tracer
		r = r.WithContext(trace.WithGraphQLRequestName(r.Context(), requestName))
		r = r.WithContext(trace.WithRequestSource(r.Context(), requestSource))

		if r.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				return err
			}

			r.Body = gzipReader

			defer gzipReader.Close()
		}

		var params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			return err
		}

		cost, costErr := graphqlbackend.EstimateQueryCost(params.Query, params.Variables)
		if costErr != nil {
			log15.Warn("estimating GraphQL cost", "error", costErr)
		}

		start := time.Now()
		response := schema.Exec(r.Context(), params.Query, params.OperationName, params.Variables)
		traceGraphQL(r, params.Query, params.OperationName, params.Variables, response.Errors, cost, costErr, isInternal, requestName, string(requestSource), start)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)

		return nil
	}
}

func traceGraphQL(r *http.Request,
	queryString string,
	operationName string,
	variables map[string]interface{},
	queryErrors []*gqlerrors.QueryError,
	cost *graphqlbackend.QueryCost,
	costErr error,
	isInternal bool,
	requestName string,
	requestSource string,
	requestStart time.Time) {

	if !honey.Enabled() || traceGraphQLQueriesSample <= 0 {
		return
	}

	duration := time.Since(requestStart)
	uid, anonymous := getUID(r)

	ev := honey.Event("graphql-cost")
	ev.SampleRate = uint(traceGraphQLQueriesSample)
	ev.AddField("query", queryString)
	ev.AddField("variables", variables)
	ev.AddField("anonymous", anonymous)
	ev.AddField("uid", uid)
	ev.AddField("operationName", operationName)
	ev.AddField("isInternal", isInternal)
	// Honeycomb has built in support for latency if you use milliseconds. We
	// multiply seconds by 1000 here instead of using d.Milliseconds() so that we
	// don't truncate durations of less than 1 millisecond.
	ev.AddField("durationMilliseconds", duration.Seconds()*1000)
	ev.AddField("hasQueryErrors", len(queryErrors) > 0)
	ev.AddField("requestName", requestName)
	ev.AddField("requestSource", requestSource)

	if costErr != nil {
		log15.Warn("estimating GraphQL cost", "error", costErr)
		ev.AddField("hasCostError", true)
		ev.AddField("costError", costErr.Error())
	} else {
		ev.AddField("hasCostError", false)
		ev.AddField("cost", cost.FieldCount)
		ev.AddField("depth", cost.MaxDepth)
		ev.AddField("costVersion", cost.Version)
	}

	_ = ev.Send()
}

var traceGraphQLQueriesSample = func() int {
	rate, _ := strconv.Atoi(os.Getenv("TRACE_GRAPHQL_QUERIES_SAMPLE"))
	return rate
}()

func getUID(r *http.Request) (uid string, anonymous bool) {
	a := actor.FromContext(r.Context())
	anonymous = !a.IsAuthenticated()
	if !anonymous {
		return a.UIDString(), anonymous
	}
	if cookie, err := r.Cookie("sourcegraphAnonymousUid"); err == nil && cookie.Value != "" {
		return cookie.Value, anonymous
	}
	// The user is anonymous with no cookie, use IP
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip, anonymous
	}
	return "unknown", anonymous
}

// guessSource guesses the source the request came from (browser, other HTTP client, etc.)
func guessSource(r *http.Request) trace.SourceType {
	userAgent := r.UserAgent()
	for _, guess := range []string{
		"Mozilla",
		"WebKit",
		"Gecko",
		"Chrome",
		"Firefox",
		"Safari",
		"Edge",
	} {
		if strings.Contains(userAgent, guess) {
			return trace.SourceBrowser
		}
	}
	return trace.SourceOther
}
