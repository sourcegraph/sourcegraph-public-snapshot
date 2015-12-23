package worker

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/auth/sharedsecret"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
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

	cl := cli.Client()

	go buildReaper(cli.Ctx)

	builders := make(chan struct{}, c.Parallel)
	for i := 0; i < c.Parallel; i++ {
		builders <- struct{}{}
	}

	for range builders {
		build, err := cl.Builds.DequeueNext(cli.Ctx, &sourcegraph.BuildsDequeueNextOp{})
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(c.DequeueMsec)))
			} else {
				log15.Error("Error dequeuing build", "err", err)
				time.Sleep(5 * time.Second)
			}
			builders <- struct{}{}
			continue
		}

		go func() {
			startBuild(cli.Ctx, build)
			builders <- struct{}{}
		}()
	}

	// TODO(sqs): mark all currently running builds as "terminated by
	// quit" before we quit (if this process is killed) so we know to rebuild them without waiting for the whole timeout period.
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

	src := cli.UpdateGlobalTokenSource{TokenSource: sharedsecret.ShortTokenSource(k, "worker:build")}
	tok, err := src.Token()
	if err != nil {
		return err
	}

	// Authenticate future requests.
	cli.Ctx = sourcegraph.WithCredentials(cli.Ctx, sharedsecret.DefensiveReuseTokenSource(tok, src))
	return nil
}
