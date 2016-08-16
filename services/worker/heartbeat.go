package worker

import (
	"log"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

// a reasonable guess; TODO(sqs): check the server's actual setting
const serverHeartbeatInterval = 15 * time.Second

// startHeartbeat spawns a background goroutine that sends heartbeats to the server until done is called.
func startHeartbeat(ctx context.Context, build sourcegraph.BuildSpec) (done func(), err error) {
	quitCh := make(chan struct{})

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		t := time.NewTicker(serverHeartbeatInterval)
		for {
			select {
			case _, ok := <-t.C:
				if !ok {
					return
				}
				now := pbtypes.NewTimestamp(time.Now())
				_, err := cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{Build: build, Info: sourcegraph.BuildUpdate{HeartbeatAt: &now}})
				if err != nil {
					log.Printf("Worker heartbeat failed in Builds.Update call for build %+v: %s.", build, err)
					return
				}
			case <-quitCh:
				t.Stop()
				return
			}
		}
	}()

	return func() {
		close(quitCh)
	}, nil
}
