package sgx

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/grpccache"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/cacheutil"
)

func init() {
	c, err := cli.CLI.AddCommand("push",
		"upload and import the current commit (to a remote)",
		"The push command reads build data from the local .srclib-cache directory and imports it into a remote Sourcegraph server. It allows users to run srclib locally (instead of, e.g., by triggering a build on the server) and see the results on a remote Sourcegraph.",
		&pushCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	if lrepo, err := srclib.OpenLocalRepo(); err == nil {
		srclib.SetOptionDefaultValue(c.Group, "commit", lrepo.CommitID)
	}
}

type pushCmd struct {
	CommitID string `long:"commit" description:"commit ID of data to import"`
}

func (c *pushCmd) Execute(args []string) error {
	cl := Client()

	rrepo, err := getRemoteRepo(cl)
	if err != nil {
		return err
	}

	lrepo, err := srclib.OpenLocalRepo()
	if err != nil {
		return err
	}

	commitID := lrepo.CommitID
	if c.CommitID != "" {
		commitID = c.CommitID
	}

	repoSpec := sourcegraph.RepoSpec{URI: rrepo.URI}
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repoSpec, Rev: commitID}

	appURL, err := getRemoteAppURL(cliCtx)
	if err != nil {
		return err
	}

	if err := c.do(cliCtx, repoRevSpec); err != nil {
		return err
	}

	u, err := router.Rel.URLToRepoRev(repoRevSpec.URI, repoRevSpec.Rev)
	if err != nil {
		return err
	}
	log.Printf("# Success! View the repository at: %s", appURL.ResolveReference(u))

	return nil
}

func (c *pushCmd) do(ctx context.Context, repoRevSpec sourcegraph.RepoRevSpec) (err error) {
	cl := Client()

	// Resolve to the full commit ID, and ensure that the remote
	// server knows about the commit.
	commit, err := cl.Repos.GetCommit(ctx, &repoRevSpec)
	if err != nil {
		return err
	}
	repoRevSpec.CommitID = string(commit.ID)

	if globalOpt.Verbose {
		log.Printf("Pushing build of %s at commit %s to server at %s...", repoRevSpec.URI, repoRevSpec.CommitID, sourcegraph.GRPCEndpoint(ctx))
	}

	// Create build (with Queued=false, since this process is
	// responsible for it and we don't want the server doing
	// anything).
	build, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{RepoRev: repoRevSpec, Opt: &sourcegraph.BuildCreateOptions{
		BuildConfig: sourcegraph.BuildConfig{
			Import: true,
			Queue:  false,
		},
		Force: true,
	}})
	if err != nil {
		return err
	}
	if globalOpt.Verbose {
		log.Printf("Created build: %s", build.Spec().IDString())
	}

	// Mark the build as started.
	now := pbtypes.NewTimestamp(time.Now())
	host := fmt.Sprintf("local (user: %s, src version: %s)", os.Getenv("USER"), buildvar.Version)
	_, err = cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
		Build: build.Spec(),
		Info:  sourcegraph.BuildUpdate{StartedAt: &now, Host: host},
	})
	if err != nil {
		return err
	}

	// Create an import task in the build (with Queued=false, since as
	// with the build this process is responsible for it and we don't
	// want the server doing anything).
	importTask := &sourcegraph.BuildTask{
		CommitID: build.CommitID,
		Attempt:  build.Attempt,
		Repo:     repoRevSpec.URI,
		Op:       sourcegraph.ImportTaskOp,
	}
	tasks, err := cl.Builds.CreateTasks(ctx, &sourcegraph.BuildsCreateTasksOp{
		Build: build.Spec(),
		Tasks: []*sourcegraph.BuildTask{importTask},
	})
	if err != nil {
		return err
	}
	importTask = tasks.BuildTasks[0]
	if globalOpt.Verbose {
		log.Printf("Created import task #%d", importTask.TaskID)
	}

	// Mark import task as started.
	now = pbtypes.NewTimestamp(time.Now())
	_, err = cl.Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
		Task: importTask.Spec(),
		Info: sourcegraph.TaskUpdate{StartedAt: &now},
	})

	// Stream logs.
	done := make(chan struct{})
	go c.streamTaskLogs(ctx, importTask.Spec(), done)
	defer func() {
		done <- struct{}{}
	}()

	// Update the task and build statuses after completion (whether
	// success or failure).
	defer func() {
		now := pbtypes.NewTimestamp(time.Now())
		_, err2 := cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
			Build: build.Spec(),
			Info:  sourcegraph.BuildUpdate{EndedAt: &now, Success: err == nil, Failure: err != nil},
		})
		if err2 != nil {
			msg := fmt.Sprintf("updating build after completion: %s", err2)
			if err != nil {
				msg += fmt.Sprintf(" (underlying error: %v)", err)
			}
			err = errors.New(msg)
		}
	}()
	defer func() {
		now := pbtypes.NewTimestamp(time.Now())
		_, err2 := cl.Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
			Task: importTask.Spec(),
			Info: sourcegraph.TaskUpdate{EndedAt: &now, Success: err == nil, Failure: err != nil},
		})
		if err2 != nil {
			msg := fmt.Sprintf("updating task after completion: %s", err2)
			if err != nil {
				msg += fmt.Sprintf(" (underlying error: %v)", err)
			}
			err = errors.New(msg)
		}
	}()

	// Perform the import.
	srcstore := pb.Client(ctx, pb.NewMultiRepoImporterClient(cl.Conn))

	bdfs, err := srclib.GetBuildDataFS(build.CommitID)
	if err != nil {
		return fmt.Errorf("getting local build data FS for %s@%s: %s", repoRevSpec.URI, build.CommitID, err)
	}

	// Importing doesn't require actual toolchains to be present (or
	// any toolchain-specific logic).
	toolchain.NoToolchains = true

	importOpt := srclib.ImportOpt{
		Repo:     repoRevSpec.URI,
		Unit:     importTask.Unit,
		UnitType: importTask.UnitType,
		CommitID: build.CommitID,
		Verbose:  globalOpt.Verbose,
	}
	if err := srclib.Import(bdfs, srcstore, importOpt); err != nil {
		return fmt.Errorf("import failed: %s", err)
	}

	// Precache root dir
	appURL, err := getRemoteAppURL(cliCtx)
	if err != nil {
		return err
	}
	cacheutil.HTTPAddr = appURL.String()
	cacheutil.PrecacheRoot(build.Repo)

	return nil
}

func (c *pushCmd) streamTaskLogs(ctx context.Context, task sourcegraph.TaskSpec, done <-chan struct{}) {
	cl := Client()

	var logOpt sourcegraph.BuildGetLogOptions
	loopsSinceLastLog := 0
	for {
		select {
		case <-done:
			return
		case <-time.After(time.Duration(loopsSinceLastLog+1) * 500 * time.Millisecond):
			logs, err := cl.Builds.GetTaskLog(
				grpccache.NoCache(ctx),
				&sourcegraph.BuildsGetTaskLogOp{Task: task, Opt: &logOpt},
			)
			if err != nil {
				if grpc.Code(err) != codes.Unimplemented {
					log.Printf("Warning: failed to get build logs: %s.", err)
				}
				<-done
				return
			}
			if len(logs.Entries) == 0 {
				loopsSinceLastLog++
				continue
			}
			logOpt.MinID = logs.MaxID
			for _, e := range logs.Entries {
				fmt.Println(e)
			}
			loopsSinceLastLog = 0
		}
	}
}
