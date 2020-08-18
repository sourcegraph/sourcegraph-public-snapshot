package heartbeat

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/efritz/glock"
	"github.com/google/go-cmp/cmp"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	queuemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client/mocks"
)

func TestHeartbeat(t *testing.T) {
	queueClient := queuemocks.NewMockClient()
	indexManager := indexmanager.New()
	clock := glock.NewMockClock()
	options := HeartbeaterOptions{
		Interval: time.Second,
	}

	indexManager.AddID(1)
	indexManager.AddID(2)
	indexManager.AddID(4)
	indexManager.AddID(5)

	heartbeater := newHeartbeater(context.Background(), queueClient, indexManager, options, clock)
	go func() { heartbeater.Start() }()
	clock.BlockingAdvance(time.Second)
	heartbeater.Stop()

	if callCount := len(queueClient.HeartbeatFunc.History()); callCount < 1 {
		t.Errorf("unexpected heartbeat call count. want>=%d have=%d", 1, callCount)
	} else {
		ids := queueClient.HeartbeatFunc.History()[0].Arg1
		sort.Ints(ids)

		if diff := cmp.Diff([]int{1, 2, 4, 5}, ids); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
	}
}
