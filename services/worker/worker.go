package worker

import (
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// a token for the worker that fits to the default id key and does not expire
var defaultAccessToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJTY29wZSI6IndvcmtlcjpidWlsZCJ9.j1SZum6h_RqEclMiQtU9KcLZoqjhyXCa-pOPynBFAUg"

// RunWorker starts the worker loop with the given parameters.
func RunWorker(ctx context.Context, endpoint *url.URL, parallel int, dequeueMsec int) error {
	ctx = sourcegraph.WithGRPCEndpoint(ctx, endpoint)
	accessToken := os.Getenv("SRC_ACCESS_TOKEN")
	if accessToken == "" {
		accessToken = defaultAccessToken
	}
	ctx = sourcegraph.WithAccessToken(ctx, accessToken)

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)

	go buildReaper(ctx)

	// Watch for sigkill so we can mark builds as ended before termination.
	var (
		wg           sync.WaitGroup
		activeBuilds = newActiveBuilds()
		killc        = make(chan os.Signal, 1)
	)
	signal.Notify(killc, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-killc
		activeBuilds.RLock()
		defer activeBuilds.RUnlock()
		cancel()
		buildCleanup(ctx, activeBuilds)
	}()

	throttle := time.Tick(time.Second / time.Duration(parallel))

	builders := make(chan struct{}, parallel)
	for i := 0; i < parallel; i++ {
		builders <- struct{}{}
	}

	for range builders {
		<-throttle // rate limit our calls to DequeueNext
		build, err := cl.Builds.DequeueNext(ctx, &sourcegraph.BuildsDequeueNextOp{})
		if err != nil {
			if grpc.Code(err) == codes.Canceled {
				break
			}
			if grpc.Code(err) == codes.NotFound {
				time.Sleep(time.Millisecond * time.Duration(dequeueMsec/2+rand.Intn(dequeueMsec)))
			} else {
				log15.Error("Error dequeuing build", "err", err)
				time.Sleep(5 * time.Second)
			}
			builders <- struct{}{}
			continue
		}
		activeBuilds.Add(build)

		wg.Add(1)
		go func() {
			defer wg.Done()
			startBuild(ctx, build)
			activeBuilds.Remove(build)
			builders <- struct{}{}
		}()
	}

	wg.Wait()
	os.Exit(1)
	panic("unreachable")
}

type activeBuilds struct {
	sync.RWMutex
	Builds map[*sourcegraph.BuildJob]struct{}
}

func newActiveBuilds() *activeBuilds {
	return &activeBuilds{
		Builds: map[*sourcegraph.BuildJob]struct{}{},
	}
}

func (a *activeBuilds) Add(build *sourcegraph.BuildJob) {
	a.Lock()
	a.Builds[build] = struct{}{}
	a.Unlock()
}

func (a *activeBuilds) Remove(build *sourcegraph.BuildJob) {
	a.Lock()
	delete(a.Builds, build)
	a.Unlock()
}
