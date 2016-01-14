package worker

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/auth/sharedsecret"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/sgx/client"
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
	Parallel    int  `short:"p" long:"parallel" description:"number of parallel builds to run" default:"2"`
	DequeueMsec int  `long:"dequeue-msec" description:"if no builds are dequeued, sleep up to this many msec before trying again" default:"1000"`
	Remote      bool `long:"remote" description:"run worker remotely from server; worker is authenticated via a shared secret token derived from the server's id key (SRC_ID_KEY_DATA)" default:"false"`
}

func (c *WorkCmd) Execute(args []string) error {
	if c.Parallel <= 0 {
		return errors.New("-p/--parallel must be > 0")
	}
	if c.Remote {
		if err := c.authenticateWorkerCtx(); err != nil {
			return fmt.Errorf("authenticating remote worker failed: %s", err)
		}
	}

	cl := client.Client()
	ctx := client.Ctx

	go buildReaper(ctx)

	// Watch for sigkill so we can mark builds as ended before termination.
	var (
		activeBuildsMu sync.Mutex
		activeBuilds   = map[*sourcegraph.Build]struct{}{}
	)
	killc := make(chan os.Signal, 1)
	signal.Notify(killc, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-killc

		// Mark all active builds (and their tasks) as killed. But set
		// an aggressive timeout so we don't block the termination for
		// too long.
		activeBuildsMu.Lock()
		defer activeBuildsMu.Unlock()
		cancel()
		if len(activeBuilds) == 0 {
			return
		}
		ctx, cancel2 := context.WithTimeout(client.Ctx, 1*time.Second)
		defer cancel2()
		time.AfterFunc(500*time.Millisecond, func() {
			// Log if it's taking a noticeable amount of time.
			builds := make([]string, 0, len(activeBuilds))
			for b := range activeBuilds {
				builds = append(builds, b.Spec().IDString())
			}
			log15.Info("Marking active builds as killed before terminating...", "builds", builds)
		})
		for b := range activeBuilds {
			if err := markBuildAsKilled(ctx, b.Spec()); err != nil {
				log15.Error("Error marking build as killed upon process termination", "build", b.Spec(), "err", err)
			}
		}
	}()

	throttle := time.Tick(time.Second / time.Duration(c.Parallel))

	builders := make(chan struct{}, c.Parallel)
	for i := 0; i < c.Parallel; i++ {
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
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(c.DequeueMsec)))
			} else {
				log15.Error("Error dequeuing build", "err", err)
				time.Sleep(5 * time.Second)
			}
			builders <- struct{}{}
			continue
		}

		// Add to active list.
		activeBuildsMu.Lock()
		activeBuilds[build] = struct{}{}
		activeBuildsMu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			startBuild(ctx, build)

			// Remove from active list.
			activeBuildsMu.Lock()
			delete(activeBuilds, build)
			activeBuildsMu.Unlock()

			builders <- struct{}{}
		}()
	}

	wg.Wait()
	os.Exit(1)
	panic("unreachable")
}

func (c *WorkCmd) authenticateWorkerCtx() error {
	idKeyData := os.Getenv("SRC_ID_KEY_DATA")
	if idKeyData == "" {
		return errors.New("SRC_ID_KEY_DATA is not set")
	}

	k, err := idkey.FromString(idKeyData)
	if err != nil {
		return err
	}

	src := client.UpdateGlobalTokenSource{TokenSource: sharedsecret.ShortTokenSource(k, "worker:build")}
	tok, err := src.Token()
	if err != nil {
		return err
	}

	// Authenticate future requests.
	client.Ctx = sourcegraph.WithCredentials(client.Ctx, sharedsecret.DefensiveReuseTokenSource(tok, src))
	return nil
}

func markBuildAsKilled(ctx context.Context, b sourcegraph.BuildSpec) error {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
		Build: b,
		Info: sourcegraph.BuildUpdate{
			EndedAt: now(),
			Killed:  true,
		},
	})
	if err != nil {
		return err
	}

	// Mark all of the build's unfinished tasks as failed, too.
	for page := int32(1); ; page++ {
		tasks, err := cl.Builds.ListBuildTasks(ctx, &sourcegraph.BuildsListBuildTasksOp{
			Build: b,
			Opt:   &sourcegraph.BuildTaskListOptions{ListOptions: sourcegraph.ListOptions{Page: page}},
		})
		if err != nil {
			return err
		}

		for _, task := range tasks.BuildTasks {
			if task.EndedAt != nil {
				continue
			}
			_, err := cl.Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
				Task: task.Spec(),
				Info: sourcegraph.TaskUpdate{Failure: true, EndedAt: now()},
			})
			if err != nil {
				return err
			}
		}
		if len(tasks.BuildTasks) == 0 {
			break
		}
	}

	return nil
}
