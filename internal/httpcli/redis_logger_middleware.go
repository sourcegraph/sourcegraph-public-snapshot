package httpcli

import (
	"encoding/json"
	"fmt"
	"github.com/go-stack/stack"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"io"
	"net/http"
	"strings"
	"time"
)

const keyPrefix = "outbound:"

const N = 50

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
			errorMessage := ""
			if err != nil {
				errorMessage = err.Error()
			}
			logItem := types.OutboundRequestLogItem{
				StartedAt:          start,
				Method:             req.Method,
				URL:                req.URL.String(),
				RequestHeaders:     removeSensitiveHeaders(req.Header),
				RequestBody:        string(requestBody),
				StatusCode:         int32(resp.StatusCode),
				ResponseHeaders:    removeSensitiveHeaders(resp.Header),
				Duration:           duration.Seconds(),
				ErrorMessage:       errorMessage,
				CreationStackFrame: creatorStack,
				CallStackFrame:     callStack.String(),
			}

			logItemJson, jsonErr := json.Marshal(logItem)

			if jsonErr != nil {
				log15.Error("RedisLoggerMiddleware failed to marshal JSON", "error", jsonErr)
			}

			// Save new item
			redisCache.Set(generateKey(time.Now()), logItemJson)

			// Delete excess items
			deletionErr := deleteOldKeys()
			if deletionErr != nil {
				log15.Error("RedisLoggerMiddleware failed to delete old keys", "error", deletionErr)
			}

			return resp, err
		})
	}
}

func generateKey(now time.Time) string {
	return fmt.Sprintf("%s%s", keyPrefix, now.UTC().Format("2006-01-02T15_04_05.999999999"))
}

func deleteOldKeys() error {
	keys, err := redisCache.GetAll(keyPrefix)
	if err != nil {
		return err
	}

	if len(keys) > N {
		// Delete all but the last N keys
		excessKeys := keys[:len(keys)-N]
		for _, key := range excessKeys {
			redisCache.Delete(key)
		}
	}
	return nil
}

func GetAllOutboundRequestLogItems() ([]*types.OutboundRequestLogItem, error) {
	rawItems, err := getAllOutboundRequestRawValues()
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

func getAllOutboundRequestRawValues() ([][]byte, error) {
	keys, err := redisCache.GetAll(keyPrefix)
	if err != nil {
		return nil, err
	}

	// Limit to N
	if len(keys) > N {
		keys = keys[len(keys)-N:]
	}

	return redisCache.GetMulti(keys...), nil
}

func removeSensitiveHeaders(headers http.Header) http.Header {
	var cleanHeaders = make(http.Header)
	for key, value := range headers {
		if isRiskyKey(key) || hasRiskyValue(value) {
			cleanHeaders[key] = []string{"REDACTED"}
		} else {
			cleanHeaders[key] = value
		}
	}
	return cleanHeaders
}

func isRiskyKey(key string) bool {
	riskyHeaderKeys := []string{"auth", "cookie", "token"}
	for _, riskyKey := range riskyHeaderKeys {
		if strings.Contains(strings.ToLower(key), riskyKey) {
			return true
		}
	}
	return false
}

func hasRiskyValue(values []string) bool {
	riskyHeaderValues := []string{"bearer", "ghp_", "glpat-"}
	for _, value := range values {
		for _, riskyValue := range riskyHeaderValues {
			if strings.Contains(strings.ToLower(value), riskyValue) {
				return true
			}
		}
	}
	return false
}
