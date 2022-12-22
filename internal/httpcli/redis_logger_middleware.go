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

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var outboundRequestsRedisCache = rcache.NewWithTTL("outbound-requests", 604800)

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
			if req != nil && req.Body != nil {
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

			// Save new item
			outboundRequestsRedisCache.Set(key, logItemJson)

			// Delete excess items
			if deleteErr := deleteExcessItems(outboundRequestsRedisCache, int(limit)); deleteErr != nil {
				middlewareErrors = errors.Append(middlewareErrors,
					errors.Wrap(deleteErr, "delete excess items"))
			}

			return resp, err
		})
	}
}

func deleteExcessItems(c *rcache.Cache, limit int) error {
	keys, err := c.ListKeys(context.Background())
	if err != nil {
		return errors.Wrap(err, "list keys")
	}

	// Delete all but the last N keys
	if len(keys) > limit {
		sort.Strings(keys)
		c.DeleteMulti(keys[:len(keys)-limit]...)
	}

	return nil
}

// GetOutboundRequestLogItems returns all outbound request log items after the given key,
// in ascending order, trimmed to maximum {limit} items. Example for `after`: "2021-01-01T00_00_00.000000".
func GetOutboundRequestLogItems(ctx context.Context, after string) ([]*types.OutboundRequestLogItem, error) {
	var limit = OutboundRequestLogLimit()

	if limit == 0 {
		return []*types.OutboundRequestLogItem{}, nil
	}

	// Get values from Redis
	rawItems, err := getAllValuesAfter(ctx, outboundRequestsRedisCache, after, int(limit))
	if err != nil {
		return nil, err
	}

	// Convert raw Redis store items to log items
	items := make([]*types.OutboundRequestLogItem, 0, len(rawItems))
	for _, rawItem := range rawItems {
		var item types.OutboundRequestLogItem
		err = json.Unmarshal(rawItem, &item)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func GetOutboundRequestLogItem(key string) (*types.OutboundRequestLogItem, error) {
	rawItem, ok := outboundRequestsRedisCache.Get(key)
	if !ok {
		return nil, errors.New("item not found")
	}

	var item types.OutboundRequestLogItem
	err := json.Unmarshal(rawItem, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// getAllValuesAfter returns all items after the given key, in ascending order, trimmed to maximum {limit} items.
func getAllValuesAfter(ctx context.Context, c *rcache.Cache, after string, limit int) ([][]byte, error) {
	all, err := c.ListKeys(ctx)
	if err != nil {
		return nil, err
	}

	var keys []string
	if after != "" {
		for _, key := range all {
			if key > after {
				keys = append(keys, key)
			}
		}
	} else {
		keys = all
	}

	// Sort ascending
	sort.Strings(keys)

	// Limit to N
	if len(keys) > limit {
		keys = keys[len(keys)-limit:]
	}

	return c.GetMulti(keys...), nil
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
