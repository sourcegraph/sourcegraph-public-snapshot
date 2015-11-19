package sgx

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/net/context"

	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

type runBuildCmd struct {
	prepBuildCmd
	Clean bool `long:"clean" description:"remove temp dir and build data after build (regardless of success/failure)"`
}

func (c *runBuildCmd) Execute(args []string) error {
	if err := c.prepBuildCmd.Execute(nil); err != nil {
		return err
	}

	if err := os.Chdir(c.BuildDir); err != nil {
		return err
	}
	doBuild := &doBuildCmd{
		Attempt:  c.Attempt,
		CommitID: c.CommitID,
		Repo:     c.Repo,
	}
	if err := doBuild.Execute(nil); err != nil {
		return err
	}

	if c.Clean {
		if srclibUseDockerExeMethod() {
			// Must use Docker to remove directory contents because if any
			// files were created in the PreConfigCmds, they will be owned
			// by root and the current user can't necessarily delete them.
			if strings.Contains(c.BuildDir, ":") {
				panic("BuildDir contains ':': " + c.BuildDir + " (could be misinterpreted in docker run --volume)")
			}
			cmd := exec.Command("docker", "run", "--volume="+filepath.Dir(c.BuildDir)+":/tmp/build-dir-parent", "--rm", "--entrypoint=/bin/rm", "ubuntu:14.04", "-rf", "--one-file-system", filepath.Join("/tmp/build-dir-parent", filepath.Base(c.BuildDir)))
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}

		}
		if err := os.RemoveAll(c.BuildDir); err != nil {
			return err
		}
	}

	return nil
}

// buildHeartbeat sends heartbeats to the build DB until c is closed.
func workerHeartbeat(ctx context.Context, bs sourcegraph.BuildsClient, interval time.Duration, build sourcegraph.BuildSpec, c <-chan struct{}) {
	t := time.NewTicker(interval)
	for {
		select {
		case _, ok := <-t.C:
			if !ok {
				return
			}
			now := pbtypes.NewTimestamp(time.Now())
			_, err := bs.Update(ctx, &sourcegraph.BuildsUpdateOp{Build: build, Info: sourcegraph.BuildUpdate{HeartbeatAt: &now}})
			if err != nil {
				log.Printf("Worker heartbeat failed in BuildsService.Update call for build %+v: %s.", build, err)
				return
			}
		case <-c:
			t.Stop()
			return
		}
	}
}
