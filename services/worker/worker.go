package worker

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/sharedsecret"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/client"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func init() {
	_, err := cli.CLI.AddCommand("work",
		"worker",
		`
Runs the worker, which monitors the build and other queues and spawns processes to run
builds.`,
		&WorkCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type WorkCmd struct {
	Parallel    int `short:"p" long:"parallel" description:"number of parallel builds to run" default:"2" env:"SRC_WORK_PARALLEL"`
	DequeueMsec int `long:"dequeue-msec" description:"if no builds are dequeued, sleep up to this many msec before trying again" default:"1000" env:"SRC_WORK_DEQUEUE_MSEC"`
}

func (c *WorkCmd) Execute(args []string) error {
	if c.Parallel <= 0 {
		return errors.New("-p/--parallel must be > 0")
	}

	idKeyData := os.Getenv("SRC_ID_KEY_DATA")
	if idKeyData == "" {
		return errors.New("SRC_ID_KEY_DATA is not available")
	}
	key, err := idkey.FromString(idKeyData)
	if err != nil {
		return err
	}

	return RunWorker(key, c.Parallel, c.DequeueMsec)
}

// RunWorker starts the worker loop with the given parameters.
func RunWorker(key *idkey.IDKey, parallel int, dequeueMsec int) error {
	client.Ctx = sourcegraph.WithCredentials(client.Ctx, sharedsecret.DefensiveReuseTokenSource(nil, sharedsecret.ShortTokenSource(key, "worker:build")))

	cl := client.Client()
	ctx, cancel := context.WithCancel(client.Ctx)

	go buildReaper(client.Ctx)

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
		buildCleanup(client.Ctx, activeBuilds)
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
