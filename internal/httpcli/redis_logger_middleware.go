package httpcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// outboundRequestsRedisFIFOList is a FIFO redis cache to store the requests.
var outboundRequestsRedisFIFOList = rcache.NewFIFOListDynamic("outbound-requests", func() int {
	return int(OutboundRequestLogLimit())
})

const sourcegraphPrefix = "github.com/sourcegraph/sourcegraph/"

func redisLoggerMiddleware() Middleware {
	creatorStackFrame, _ := getFrames(4).Next()
	return func(cli Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			resp, err := cli.Do(req)
			duration := time.Since(start)

			limit := OutboundRequestLogLimit()
			shouldRedactSensitiveHeaders := !deploy.IsDev(deploy.Type()) || RedactOutboundRequestHeaders()

			// Feature is turned off, do not log
			if limit == 0 {
				return resp, err
			}

			// middlewareErrors will be set later if there is an error
			var middlewareErrors error
			defer func() {
				if middlewareErrors != nil {
					*req = *req.WithContext(context.WithValue(req.Context(),
						redisLoggingMiddlewareErrorKey, middlewareErrors))
				}
			}()

			// Read body
			var requestBody []byte
			if req != nil && req.GetBody != nil {
				body, _ := req.GetBody()
				if body != nil {
					var readErr error
					requestBody, readErr = io.ReadAll(body)
					if err != nil {
						middlewareErrors = errors.Append(middlewareErrors,
							errors.Wrap(readErr, "read body"))
					}
				}
			}

			// Pull out data if we have `resp`
			var responseHeaders http.Header
			var statusCode int32
			if resp != nil {
				responseHeaders = resp.Header
				statusCode = int32(resp.StatusCode)
			}

			// Redact sensitive headers
			requestHeaders := req.Header

			if shouldRedactSensitiveHeaders {
				requestHeaders = redactSensitiveHeaders(requestHeaders)
				responseHeaders = redactSensitiveHeaders(responseHeaders)
			}

			// Create log item
			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}
			key := time.Now().UTC().Format("2006-01-02T15_04_05.999999999")
			callerStackFrames := getFrames(4) // Starts at the caller of the caller of redisLoggerMiddleware
			logItem := types.OutboundRequestLogItem{
				ID:                 key,
				StartedAt:          start,
				Method:             req.Method,
				URL:                req.URL.String(),
				RequestHeaders:     requestHeaders,
				RequestBody:        string(requestBody),
				StatusCode:         statusCode,
				ResponseHeaders:    responseHeaders,
				Duration:           duration.Seconds(),
				ErrorMessage:       errorMessage,
				CreationStackFrame: formatStackFrame(creatorStackFrame.Function, creatorStackFrame.File, creatorStackFrame.Line),
				CallStackFrame:     formatStackFrames(callerStackFrames),
			}

			// Serialize log item
			logItemJson, jsonErr := json.Marshal(logItem)
			if jsonErr != nil {
				middlewareErrors = errors.Append(middlewareErrors,
					errors.Wrap(jsonErr, "marshal log item"))
			}

			go func() {
				// Save new item
				if err := outboundRequestsRedisFIFOList.Insert(logItemJson); err != nil {
					// Log would get upset if we created a logger at init time → create logger on the fly
					log.Scoped("redisLoggerMiddleware").Error("insert log item", log.Error(err))
				}
			}()

			return resp, err
		})
	}
}

// GetOutboundRequestLogItems returns all outbound request log items after the given key,
// in ascending order, trimmed to maximum {limit} items. Example for `after`: "2021-01-01T00_00_00.000000".
func GetOutboundRequestLogItems(ctx context.Context, after string) ([]*types.OutboundRequestLogItem, error) {
	var limit = int(OutboundRequestLogLimit())

	if limit == 0 {
		return []*types.OutboundRequestLogItem{}, nil
	}

	items, err := getOutboundRequestLogItems(ctx, func(item *types.OutboundRequestLogItem) bool {
		if after == "" {
			return true
		} else {
			return item.ID > after
		}
	})
	if err != nil {
		return nil, err
	}

	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

func GetOutboundRequestLogItem(key string) (*types.OutboundRequestLogItem, error) {
	items, err := getOutboundRequestLogItems(context.Background(), func(item *types.OutboundRequestLogItem) bool {
		return item.ID == key
	})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("item not found")
	}
	return items[0], nil
}

// getOutboundRequestLogItems returns all items where pred returns true,
// sorted by ID ascending.
func getOutboundRequestLogItems(ctx context.Context, pred func(*types.OutboundRequestLogItem) bool) ([]*types.OutboundRequestLogItem, error) {
	// We fetch all values from redis, then just return those matching pred.
	// Given the max size is enforced as 500, this is fine. But if we ever
	// raise the limit, we likely need to think of an alternative way to do
	// pagination against lists / or also store the items so we can look up by
	// key
	rawItems, err := outboundRequestsRedisFIFOList.All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list all log items")
	}

	var items []*types.OutboundRequestLogItem
	for _, rawItem := range rawItems {
		var item types.OutboundRequestLogItem
		err = json.Unmarshal(rawItem, &item)
		if err != nil {
			return nil, err
		}
		if pred(&item) {
			items = append(items, &item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

func redactSensitiveHeaders(headers http.Header) http.Header {
	var cleanHeaders = make(http.Header)
	for name, values := range headers {
		if IsRiskyHeader(name, values) {
			cleanHeaders[name] = []string{"REDACTED"}
		} else {
			cleanHeaders[name] = values
		}
	}
	return cleanHeaders
}

func formatStackFrames(frames *runtime.Frames) string {
	var sb strings.Builder
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		sb.WriteString(formatStackFrame(frame.Function, frame.File, frame.Line))
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formatStackFrame(function string, file string, line int) string {
	treeAndFunc := strings.Split(function, "/")   // github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend.(*requestTracer).TraceQuery
	pckAndFunc := treeAndFunc[len(treeAndFunc)-1] // graphqlbackend.(*requestTracer).TraceQuery
	dotPieces := strings.Split(pckAndFunc, ".")   // ["graphqlbackend" , "(*requestTracer)", "TraceQuery"]
	pckName := dotPieces[0]                       // graphqlbackend
	funcName := strings.Join(dotPieces[1:], ".")  // (*requestTracer).TraceQuery

	tree := strings.Join(treeAndFunc[:len(treeAndFunc)-1], "/") + "/" + pckName // github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend
	tree = strings.TrimPrefix(tree, sourcegraphPrefix)

	// Reconstruct the frame file path so that we don't include the local path on the machine that built this instance
	fileName := strings.TrimPrefix(filepath.Join(tree, filepath.Base(file)), "/main/") // cmd/frontend/graphqlbackend/trace.go

	return fmt.Sprintf("%s:%d (Function: %s)", fileName, line, funcName)
}

const pcLen = 1024

func getFrames(skip int) *runtime.Frames {
	pc := make([]uintptr, pcLen)
	n := runtime.Callers(skip, pc)
	return runtime.CallersFrames(pc[:n])
}
