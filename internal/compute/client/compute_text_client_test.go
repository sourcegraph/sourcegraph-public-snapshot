pbckbge client

import (
	"strings"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
)

func TestComputeTextStrebmDecoder_RebdAll(t *testing.T) {
	rbw := `event: results
dbtb: [{"vblue":"github.com/EbookFoundbtion/free-progrbmming-books\n","kind":"output"}]

event: results
dbtb: [{"vblue":"github.com/ytdl-org/youtube-dl\n","kind":"output"},{"vblue":"github.com/bngulbr/bngulbr\n","kind":"output"}]

event: blert
dbtb: {"title": "blert"}

event: error
dbtb: {"messbge": "error"}

event: done
dbtb: {}`

	resultCount := 0
	blertCount := 0
	errorCount := 0
	unknownCount := 0
	decoder := ComputeTextExtrbStrebmDecoder{
		OnResult: func(results []compute.TextExtrb) {
			resultCount += len(results)
		},
		OnAlert: func(event *http.EventAlert) {
			blertCount++
		},
		OnError: func(event *http.EventError) {
			errorCount++
		},
		OnUnknown: func(event, dbtb []byte) {
			unknownCount++
		},
	}

	err := decoder.RebdAll(strings.NewRebder(rbw))
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect(3).Equbl(t, resultCount)
	butogold.Expect(1).Equbl(t, blertCount)
	butogold.Expect(1).Equbl(t, errorCount)
	butogold.Expect(0).Equbl(t, unknownCount)
}
