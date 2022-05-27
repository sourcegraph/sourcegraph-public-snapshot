package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/inconshreveable/log15"
	"github.com/throttled/throttled/v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func serveGraphQL(schema *graphql.Schema, rlw graphqlbackend.LimitWatcher, isInternal bool) func(w http.ResponseWriter, r *http.Request) (err error) {
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
		requestSource := search.GuessSource(r)

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

		var params graphQLQueryParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			return err
		}

		traceData := traceData{
			queryParams:   params,
			isInternal:    isInternal,
			requestName:   requestName,
			requestSource: string(requestSource),
		}

		defer func() {
			instrumentGraphQL(traceData)
			traceGraphQL(traceData)
		}()

		uid, isIP, anonymous := getUID(r)
		traceData.uid = uid
		traceData.anonymous = anonymous

		validationErrs := schema.ValidateWithVariables(params.Query, params.Variables)

		var cost *graphqlbackend.QueryCost
		var costErr error

		// Don't attempt to estimate or rate limit a request that has failed validation
		if len(validationErrs) == 0 {
			cost, costErr = graphqlbackend.EstimateQueryCost(params.Query, params.Variables)
			if costErr != nil {
				// We send errors to Honeycomb, no need to spam logs
				log15.Debug("estimating GraphQL cost", "error", costErr)
			}
			traceData.costError = costErr
			traceData.cost = cost

			if rl, enabled := rlw.Get(); enabled && cost != nil {
				limited, result, err := rl.RateLimit(uid, cost.FieldCount, graphqlbackend.LimiterArgs{
					IsIP:          isIP,
					Anonymous:     anonymous,
					RequestName:   requestName,
					RequestSource: requestSource,
				})
				if err != nil {
					log15.Error("checking GraphQL rate limit", "error", err)
					traceData.limitError = err
				} else {
					traceData.limited = limited
					traceData.limitResult = result
					if limited {
						w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
						w.WriteHeader(http.StatusTooManyRequests)
						return nil
					}
				}
			}
		}

		traceData.execStart = time.Now()
		response := schema.Exec(r.Context(), params.Query, params.OperationName, params.Variables)
		traceData.queryErrors = response.Errors
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)

		return nil
	}
}

type graphQLQueryParams struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

type traceData struct {
	queryParams   graphQLQueryParams
	execStart     time.Time
	uid           string
	anonymous     bool
	isInternal    bool
	requestName   string
	requestSource string
	queryErrors   []*gqlerrors.QueryError

	cost      *graphqlbackend.QueryCost
	costError error

	limited     bool
	limitError  error
	limitResult throttled.RateLimitResult
}

func traceGraphQL(data traceData) {
	if !honey.Enabled() || traceGraphQLQueriesSample <= 0 {
		return
	}

	duration := time.Since(data.execStart)

	ev := honey.NewEvent("graphql-cost")
	ev.SetSampleRate(uint(traceGraphQLQueriesSample))

	ev.AddField("query", data.queryParams.Query)
	ev.AddField("variables", data.queryParams.Variables)
	ev.AddField("operationName", data.queryParams.OperationName)

	ev.AddField("anonymous", data.anonymous)
	ev.AddField("uid", data.uid)
	ev.AddField("isInternal", data.isInternal)
	// Honeycomb has built in support for latency if you use milliseconds. We
	// multiply seconds by 1000 here instead of using d.Milliseconds() so that we
	// don't truncate durations of less than 1 millisecond.
	ev.AddField("durationMilliseconds", duration.Seconds()*1000)
	ev.AddField("hasQueryErrors", len(data.queryErrors) > 0)
	ev.AddField("requestName", data.requestName)
	ev.AddField("requestSource", data.requestSource)

	if data.costError != nil {
		ev.AddField("hasCostError", true)
		ev.AddField("costError", data.costError.Error())
	} else if data.cost != nil {
		ev.AddField("hasCostError", false)
		ev.AddField("cost", data.cost.FieldCount)
		ev.AddField("depth", data.cost.MaxDepth)
		ev.AddField("costVersion", data.cost.Version)
	}

	ev.AddField("rateLimited", data.limited)
	if data.limitError != nil {
		ev.AddField("rateLimitError", data.limitError.Error())
	} else {
		ev.AddField("rateLimit", data.limitResult.Limit)
		ev.AddField("rateLimitRemaining", data.limitResult.Remaining)
	}

	_ = ev.Send()
}

var traceGraphQLQueriesSample = func() int {
	rate, _ := strconv.Atoi(os.Getenv("TRACE_GRAPHQL_QUERIES_SAMPLE"))
	return rate
}()

func getUID(r *http.Request) (uid string, ip bool, anonymous bool) {
	a := actor.FromContext(r.Context())
	anonymous = !a.IsAuthenticated()
	if !anonymous {
		return a.UIDString(), false, anonymous
	}
	if uid, ok := cookie.AnonymousUID(r); ok && uid != "" {
		return uid, false, anonymous
	}
	// The user is anonymous with no cookie, use IP
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip, true, anonymous
	}
	return "unknown", false, anonymous
}
