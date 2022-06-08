package client

import (
	"strings"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

func TestComputeTextStreamDecoder_ReadAll(t *testing.T) {
	raw := `event: results
data: [{"value":"github.com/EbookFoundation/free-programming-books\n","kind":"output"}]

event: results
data: [{"value":"github.com/ytdl-org/youtube-dl\n","kind":"output"},{"value":"github.com/angular/angular\n","kind":"output"}]

event: alert
data: {"title": "alert"}

event: error
data: {"message": "error"}

event: done
data: {}`

	resultCount := 0
	alertCount := 0
	errorCount := 0
	unknownCount := 0
	decoder := ComputeTextStreamDecoder{
		OnResult: func(results []compute.Text) {
			resultCount += len(results)
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
	autogold.Want("resultCount", int(3)).Equal(t, resultCount)
	autogold.Want("alertCount", int(1)).Equal(t, alertCount)
	autogold.Want("errorCount", int(1)).Equal(t, errorCount)
	autogold.Want("unknownCount", int(0)).Equal(t, unknownCount)
}
