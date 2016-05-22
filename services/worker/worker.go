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

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/sharedsecret"
)

// RunWorker starts the worker loop with the given parameters.
func RunWorker(key *idkey.IDKey, endpoint *url.URL, parallel int, dequeueMsec int) error {
	ctx := sourcegraph.WithGRPCEndpoint(context.Background(), endpoint)
	ctx = sourcegraph.WithCredentials(ctx, sharedsecret.DefensiveReuseTokenSource(nil, sharedsecret.ShortTokenSource(key, "worker:build")))

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
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(dequeueMsec)))
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
