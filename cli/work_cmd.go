package cli

import (
	"errors"
	"log"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker"
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
	DequeueMsec int `long:"dequeue-msec" description:"if no builds are dequeued, sleep roughly this many msec before trying again" default:"5000" env:"SRC_WORK_DEQUEUE_MSEC"`
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

	return worker.RunWorker(key, endpoint.URLOrDefault(), c.Parallel, c.DequeueMsec)
}
