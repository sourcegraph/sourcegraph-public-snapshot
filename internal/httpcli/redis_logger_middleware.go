package httpcli

import (
	"encoding/json"
	"fmt"
	"github.com/go-stack/stack"
	"github.com/inconshreveable/log15"
	"io"
	"net/http"
	"time"
)

type RedisLogItem struct {
	Method          string      `json:"method"` // The request method (GET, POST, etc.)
	URL             string      `json:"url"`
	RequestHeaders  http.Header `json:"request_headers"`
	RequestBody     string      `json:"body"`
	StatusCode      int         `json:"status_code"` // The response status code
	ResponseHeaders http.Header `json:"response_headers"`
	Duration        string      `json:"duration"`
	Error           error       `json:"error"`
	CreatedAtFrame  string      `json:"created_at_frame"`
	CalledAtFrame   string      `json:"called_at_frame"`
}

func redisLoggerMiddleware() Middleware {
	f := stack.Caller(2).Frame()
	creatorStack := fmt.Sprintf("%s:%d, %s", f.File, f.Line, f.Function)
	return func(cli Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			resp, err := cli.Do(req)
			duration := time.Since(start)
			var requestBody []byte
			if req != nil && req.Body != nil {
				requestBody, _ = io.ReadAll(req.Body)
			}
			callStack := stack.Trace().TrimRuntime().TrimBelow(stack.Caller(3))
			logItem := RedisLogItem{
				Method:          req.Method,
				URL:             req.URL.String(),
				RequestHeaders:  req.Header,
				RequestBody:     string(requestBody),
				StatusCode:      resp.StatusCode,
				ResponseHeaders: resp.Header,
				Duration:        duration.String(),
				Error:           err,
				CreatedAtFrame:  creatorStack,
				CalledAtFrame:   callStack.String(),
			}

			logItemJson, jsonErr := json.Marshal(logItem)

			if jsonErr != nil {
				log15.Error("RedisLoggerMiddleware failed to marshal JSON", "error", jsonErr)
			}

			// Current UTC date in YYYY-MM-DD format
			today := time.Now().UTC().Format("2006-01-02")
			// Current UTC time in HH:MM:SS.nS format
			now := time.Now().UTC().Format("15-04-05.999999999")

			// Redis key
			key := fmt.Sprintf("outgoing_external_requests:%s:%s", today, now)

			redisCache.Set(key, logItemJson)

			return resp, err
		})
	}
}
