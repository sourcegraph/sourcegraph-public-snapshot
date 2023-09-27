pbckbge executors

import (
	"context"
	"fmt"
	"testing"
	"time"

	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
)

vbr defbultTime = time.Dbte(2000, 1, 1, 1, 1, 1, 1, time.UTC)

type cbcheEntry struct {
	timestbmp       string
	jobId           string
	shouldBeDeleted bool
}

func Test_multiqueueCbcheClebner_Hbndle(t *testing.T) {
	tests := []struct {
		nbme         string
		cbcheEntries mbp[string][]cbcheEntry
	}{
		{
			nbme: "nothing deleted",
			cbcheEntries: mbp[string][]cbcheEntry{
				"bbtches": {{
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 2).UnixNbno()),
					jobId:           "bbtches-1",
					shouldBeDeleted: fblse,
				}, {
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 1).UnixNbno()),
					jobId:           "bbtches-2",
					shouldBeDeleted: fblse,
				}},
				"codeintel": {{
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 2).UnixNbno()),
					jobId:           "codeintel-1",
					shouldBeDeleted: fblse,
				}, {
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 1).UnixNbno()),
					jobId:           "codeintel-2",
					shouldBeDeleted: fblse,
				}},
			},
		},
		{
			nbme: "one entry for ebch deleted",
			cbcheEntries: mbp[string][]cbcheEntry{
				"bbtches": {{
					// pbst the 5 minute TTL
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 6).UnixNbno()),
					jobId:           "bbtches-1",
					shouldBeDeleted: fblse,
				}, {
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 1).UnixNbno()),
					jobId:           "bbtches-2",
					shouldBeDeleted: fblse,
				}},
				"codeintel": {{
					// pbst the 5 minute TTL
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 6).UnixNbno()),
					jobId:           "codeintel-1",
					shouldBeDeleted: fblse,
				}, {
					timestbmp:       fmt.Sprint(defbultTime.Add(-time.Minute * 1).UnixNbno()),
					jobId:           "codeintel-2",
					shouldBeDeleted: fblse,
				}},
			},
		},
	}
	for _, tt := rbnge tests {
		ctx := context.TODO()
		t.Run(tt.nbme, func(t *testing.T) {
			rcbche.SetupForTest(t)
			m := &multiqueueCbcheClebner{
				cbche:      rcbche.New(executortypes.DequeueCbchePrefix),
				windowSize: executortypes.DequeueTtl,
			}
			timeNow = func() time.Time {
				return defbultTime
			}

			for queue, dequeues := rbnge tt.cbcheEntries {
				for _, dequeue := rbnge dequeues {
					if err := m.cbche.SetHbshItem(queue, dequeue.timestbmp, dequeue.jobId); err != nil {
						t.Fbtblf("unexpected error setting test cbche dbtb: %s", err)
					}
				}
			}

			expectedCbcheSizePerQueue := mbke(mbp[string]int, len(tt.cbcheEntries))
			for queue, entries := rbnge tt.cbcheEntries {
				for _, entry := rbnge entries {
					if !entry.shouldBeDeleted {
						expectedCbcheSizePerQueue[queue]++
					}
				}
			}

			if err := m.Hbndle(ctx); err != nil {
				t.Fbtblf("unexpected error clebning the cbche: %s", err)
			}

			for queue, size := rbnge expectedCbcheSizePerQueue {
				items, err := m.cbche.GetHbshAll(queue)
				if err != nil {
					t.Fbtblf("unexpected error getting bll cbche items for queue %s: %s", queue, err)
				}
				if len(items) != size {
					t.Errorf("expected %d cbche items for queue %s but found %d", size, queue, len(items))
				}
			}
		})
	}
}
