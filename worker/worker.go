package worker

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/keegancsmith/tmpfriend"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/auth/sharedsecret"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
	"src.sourcegraph.com/sourcegraph/util/buildutil"
	"src.sourcegraph.com/sourcegraph/util/executil"
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
	DequeueMsec int    `long:"dequeue-msec" description:"if no builds are dequeued, sleep up to this many msec before trying again" default:"1000"`
	NumWorkers  int    `short:"n" long:"num-workers" description:"number of parallel workers" default:"1" env:"SG_NUM_WORKERS"`
	BuildRoot   string `long:"build-root" description:"root of dir tree in which to perform builds" default:"$SGPATH/builds"`
	Clean       bool   `long:"clean" description:"remove temp dirs and build data when the worker starts and after builds complete"`
	Remote      bool   `long:"remote" description:"run worker remotely from server; worker is authenticated via a shared secret token derived from the server's id key (SRC_ID_KEY_DATA)" default:"false"`
}

var cleanBuildData bool

func (c *WorkCmd) Execute(args []string) error {
	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	c.BuildRoot = os.ExpandEnv(c.BuildRoot)

	workersNoun := "worker"
	if c.NumWorkers != 1 {
		workersNoun += "s"
	}
	log15.Debug(fmt.Sprintf("%d %s", c.NumWorkers, workersNoun), "BuildRoot", c.BuildRoot)

	if c.Remote {
		if err := c.authenticateWorkerCtx(); err != nil {
			log.Printf("remote auth failed: %v\n", err)
			return nil
		}
	}

	cl := cli.Client()

	// Check for Docker and fail if it's required but unavailable.
	srclibUseDockerExeMethod()

	// TODO(sqs): make this slightly less than the server's
	// BuildTimeout (need to add a way for the worker client to
	// determine the BuildTimeout value).
	cmdTimeout := 89 * time.Minute

	var dequeueMu sync.Mutex
	// dequeueNext returns the next build in the queue.
	dequeueNext := func() *sourcegraph.Build {
		dequeueMu.Lock()
		defer dequeueMu.Unlock()

		build, err := cl.Builds.DequeueNext(cli.Ctx, &sourcegraph.BuildsDequeueNextOp{})
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(c.DequeueMsec)))
				return nil
			}
			log.Println("Error dequeuing build: ", err)
			return nil
		}

		return build
	}

	if c.Clean {
		cleanBuildData = true
		// Remove subdirs of build root.
		entries, err := ioutil.ReadDir(c.BuildRoot)
		if err == nil {
			for _, e := range entries {
				if err := os.RemoveAll(filepath.Join(c.BuildRoot, e.Name())); err != nil {
					return err
				}
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	// Prevent running 2+ builds of the same repo. If you do, `git checkout`
	// to the commit ID conflicts with other concurrent builds of the
	// same repo.
	var (
		buildDirsInUse   = map[string]*sync.Mutex{}
		buildDirsInUseMu sync.Mutex
	)

	for i := 1; i <= c.NumWorkers; i++ {
		i := i
		go func() {
			for {
				func() {
					build := dequeueNext()
					if build == nil {
						return
					}

					cl := cli.Client()
					ctx := cli.Ctx

					quitCh := make(chan struct{})
					defer func() {
						close(quitCh)
					}()
					// TODO(sqs): add a way for the worker client to
					// determine the server's heartbeat timeout - this
					// value is hardcoded here and this will break if
					// the server heartbeat timeout is shorter.
					const hardcodedServerHeartbeatTimeout_hack = 30 * time.Second
					go workerHeartbeat(ctx, cl.Builds, (hardcodedServerHeartbeatTimeout_hack / 2), build.Spec(), quitCh)

					buildDir := filepath.Join(c.BuildRoot, build.Repo)
					buildDirsInUseMu.Lock()
					if mu, present := buildDirsInUse[buildDir]; present {
						buildDirsInUseMu.Unlock()
						mu.Lock()
					} else {
						buildDirsInUse[buildDir] = &sync.Mutex{}
						buildDirsInUseMu.Unlock()
						buildDirsInUse[buildDir].Lock()
					}
					defer buildDirsInUse[buildDir].Unlock()

					tl := newLogger(buildutil.BuildTag(build.Spec()))
					defer tl.Close()
					lw := io.MultiWriter(os.Stderr, tl)
					blog := log.New(lw, "", 0)
					log.Printf("Starting build %s (logged in %s).", build.Spec().IDString(), tl.Destination)
					now := pbtypes.NewTimestamp(time.Now())
					build, err := cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
						Build: build.Spec(),
						Info: sourcegraph.BuildUpdate{
							StartedAt: &now,
						}})

					if err != nil {
						blog.Println("Error updating build: ", err)
						return
					}

					// Run the build (prepare the build directory, run toolchains,
					// and import data into the DB).
					cmd := cmdWithClientArgs(
						sgxcmd.Path,
						"-v", "internal-build", "run",
						"--build-dir", buildDir,
						"--commit-id", build.CommitID,
						"--repo", build.Repo,
						"--attempt", strconv.Itoa(int(build.Attempt)),
					)

					if c.Clean {
						cmd.Args = append(cmd.Args, "--clean")
					}
					if !fed.Config.IsRoot {
						// Mothership does not have private repos
						privDir := filepath.Join(c.BuildRoot, ".private_workspace_dir_"+strconv.Itoa(i))
						os.MkdirAll(filepath.Join(privDir, "src"), 0700)
						cmd.Args = append(cmd.Args, "--private-workspace-dir", privDir)
					}
					cmd.Stdout, cmd.Stderr = lw, lw
					endUpdate := sourcegraph.BuildUpdate{}
					if err := cmd.Start(); err != nil {
						log.Printf("Build #%s (%s) failed to start: %s.", build.Spec().IDString(), build.Repo, err)
						endUpdate.Success, endUpdate.Failure = false, false
					}
					if err := executil.CmdWaitWithTimeout(cmdTimeout, cmd); err == nil {
						log.Printf("Build #%s (%s) succeeded in %s.", build.Spec().IDString(), build.Repo, time.Since(build.StartedAt.Time()))
						endUpdate.Success, endUpdate.Failure = true, false
					} else if err == executil.ErrCmdTimeout {
						log.Printf("Build #%s (%s) timed out after %s.", build.Spec().IDString(), build.Repo, cmdTimeout)
						endUpdate.Success, endUpdate.Failure = false, true
					} else {
						log.Printf("Build #%s (%s) failed after %s: %s.", build.Spec().IDString(), build.Repo, time.Since(build.StartedAt.Time()), err)
						endUpdate.Success, endUpdate.Failure = false, true
					}

					now = pbtypes.NewTimestamp(time.Now())
					endUpdate.EndedAt = &now
					build, err = cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{Build: build.Spec(), Info: endUpdate})
					if err != nil {
						blog.Println("Error updating build: ", err)
					}
				}()
			}
		}()
	}

	// TODO(sqs): mark all currently running builds as "terminated by
	// quit" before we quit (if this process is killed) so we know to rebuild them without waiting for the whole timeout period.

	select {}
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
