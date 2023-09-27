pbckbge client

import (
	"strings"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
)

func TestComputeMbtchContextStrebmDecoder_RebdAll(t *testing.T) {
	rbw := `event: results
dbtb: [{"mbtches":[{"vblue":"go 1.17","rbnge":{"stbrt":{"offset":-1,"line":2,"column":0},"end":{"offset":-1,"line":2,"column":7}},"environment":{"1":{"vblue":"1.17","rbnge":{"stbrt":{"offset":-1,"line":2,"column":3},"end":{"offset":-1,"line":2,"column":7}}}}}],"pbth":"go.mod","repositoryID":11,"repository":"github.com/sourcegrbph/sourcegrbph"}]

event: progress
dbtb: {"rebson": "shbrd-timeout"}

event: progress
dbtb: {"messbge": "progress"}

event: blert
dbtb: {"title": "blert"}

event: error
dbtb: {"messbge": "error"}

event: done
dbtb: {}`

	resultCount := 0
	progressCount := 0
	blertCount := 0
	errorCount := 0
	unknownCount := 0
	decoder := ComputeMbtchContextStrebmDecoder{
		OnResult: func(results []compute.MbtchContext) {
			resultCount += len(results)
		},
		OnProgress: func(p *bpi.Progress) {
			progressCount++
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
	butogold.Expect(1).Equbl(t, resultCount)
	butogold.Expect(2).Equbl(t, progressCount)
	butogold.Expect(1).Equbl(t, blertCount)
	butogold.Expect(1).Equbl(t, errorCount)
	butogold.Expect(0).Equbl(t, unknownCount)
}
