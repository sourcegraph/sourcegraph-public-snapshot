package httpcli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/go-stack/stack"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const keyPrefix = "outbound:"

func redisLoggerMiddleware() Middleware {
	creatorStackFrame := stack.Caller(2).Frame()
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
			errorMessage := ""
			if err != nil {
				errorMessage = err.Error()
			}
			key := generateKey(time.Now())
			logItem := types.OutboundRequestLogItem{
				Key:                key,
				StartedAt:          start,
				Method:             req.Method,
				URL:                req.URL.String(),
				RequestHeaders:     removeSensitiveHeaders(req.Header),
				RequestBody:        string(requestBody),
				StatusCode:         int32(resp.StatusCode),
				ResponseHeaders:    removeSensitiveHeaders(resp.Header),
				Duration:           duration.Seconds(),
				ErrorMessage:       errorMessage,
				CreationStackFrame: formatCreatorStack(creatorStackFrame),
				CallStackFrame:     callStack.String(),
			}

			logItemJson, jsonErr := json.Marshal(logItem)

			if jsonErr != nil {
				log.Error(jsonErr)
			}

			// Save new item
			redisCache.Set(key, logItemJson)

			// Delete excess items
			deletionErr := deleteOldKeys(OutboundRequestLogLimit())
			if deletionErr != nil {
				log.Error(deletionErr)
			}

			return resp, err
		})
	}
}

func formatCreatorStack(frame runtime.Frame) string {
	functionWithoutRepoName := strings.Split(frame.Function, "/")[3:]
	return fmt.Sprintf("%s:%d, %s", frame.File, frame.Line, functionWithoutRepoName)
}

func generateKey(now time.Time) string {
	return fmt.Sprintf("%s%s", keyPrefix, now.UTC().Format("2006-01-02T15_04_05.999999999"))
}

func deleteOldKeys(limit int) error {
	keys, err := redisCache.ListKeys(nil, keyPrefix)
	if err != nil {
		return err
	}

	if len(keys) > limit {
		// Delete all but the last N keys
		sort.Strings(keys)
		excessKeys := keys[:len(keys)-limit]
		for _, key := range excessKeys {
			redisCache.Delete(key)
		}
	}
	return nil
}

func GetAllOutboundRequestLogItemsAfter(lastKey *string, limit int) ([]*types.OutboundRequestLogItem, error) {
	if limit == 0 {
		return []*types.OutboundRequestLogItem{}, nil
	}

	rawItems, err := getAllAfter(lastKey, limit)
	if err != nil {
		return nil, err
	}

	var items []*types.OutboundRequestLogItem
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

func getAllAfter(lastKey *string, limit int) ([][]byte, error) {
	all, err := redisCache.ListKeys(nil, keyPrefix)
	if err != nil {
		return nil, err
	}

	var keys []string
	if lastKey != nil {
		for _, key := range all {
			if key > *lastKey {
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

	return redisCache.GetMulti(keys...), nil
}

func removeSensitiveHeaders(headers http.Header) http.Header {
	var cleanHeaders = make(http.Header)
	for key, value := range headers {
		if IsRiskyKey(key) || HasRiskyValue(value) {
			cleanHeaders[key] = []string{"REDACTED"}
		} else {
			cleanHeaders[key] = value
		}
	}
	return cleanHeaders
}
