package graphqlbackend

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_captureSlowRequest(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		rcache.SetupForTest(t)

		ctx := context.Background()
		logger, _ := logtest.Captured(t)

		req := types.SlowRequest{
			UserID:    100,
			Name:      "Foobar",
			Source:    "Browser",
			Variables: map[string]any{"a": "b"},
			Errors:    []string{"something"},
		}

		captureSlowRequest(ctx, logger, &req)

		raws, err := slowRequestRedisRecentList.All(ctx)
		if err != nil {
			t.Errorf("expected no error, got %q", err)
		}
		if len(raws) != 1 {
			t.Errorf("expected to find one request captured, got %d", len(raws))
		}
		var got types.SlowRequest
		if err := json.Unmarshal(raws[0], &got); err != nil {
			t.Errorf("expected no error, got %q", err)
		}

		if diff := cmp.Diff(got, req); diff != "" {
			t.Errorf("request doesn't match: %s", diff)
		}
	})
}
