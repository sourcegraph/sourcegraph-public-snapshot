package client

import (
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

func TestComputeMatchContextStreamDecoder_ReadAll(t *testing.T) {
	raw := `event: results
data: [{"matches":[{"value":"go 1.17","range":{"start":{"offset":-1,"line":2,"column":0},"end":{"offset":-1,"line":2,"column":7}},"environment":{"1":{"value":"1.17","range":{"start":{"offset":-1,"line":2,"column":3},"end":{"offset":-1,"line":2,"column":7}}}}}],"path":"go.mod","repositoryID":11,"repository":"github.com/sourcegraph/sourcegraph"}]

event: progress
data: {"reason": "shard-timeout"}

event: progress
data: {"message": "progress"}

event: alert
data: {"title": "alert"}

event: error
data: {"message": "error"}

event: done
data: {}`

	resultCount := 0
	progressCount := 0
	alertCount := 0
	errorCount := 0
	unknownCount := 0
	decoder := ComputeMatchContextStreamDecoder{
		OnResult: func(results []compute.MatchContext) {
			resultCount += len(results)
		},
		OnProgress: func(p *api.Progress) {
			progressCount++
		},
		OnAlert: func(event *http.EventAlert) {
			alertCount++
		},
		OnError: func(event *http.EventError) {
			errorCount++
		},
		OnUnknown: func(event, data []byte) {
			unknownCount++
		},
	}

	err := decoder.ReadAll(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	autogold.Expect(1).Equal(t, resultCount)
	autogold.Expect(2).Equal(t, progressCount)
	autogold.Expect(1).Equal(t, alertCount)
	autogold.Expect(1).Equal(t, errorCount)
	autogold.Expect(0).Equal(t, unknownCount)
}
