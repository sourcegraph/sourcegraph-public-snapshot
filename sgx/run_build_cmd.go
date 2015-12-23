package sgx

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rogpeppe/rog-go/parallel"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

type runBuildCmd struct {
	prepBuildCmd
	Clean            bool   `long:"clean" description:"remove temp dir and build data after build (regardless of success/failure)"`
	PrivateWorkspace string `long:"private-workspace-dir" description:"ensures private repos are available in this directory before building"`
}

func (c *runBuildCmd) Execute(args []string) error {
	if c.PrivateWorkspace != "" {
		c.prepOtherRepos()
	}

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

// prepOtherRepos (HACK go-specific) is a workaround for the fact that we
// don't have creds for cloning private repos available for go get. It
// attempts to checkout all repos on the host to the default branch in build
// dir, to make them available on the GOPATH
func (c *runBuildCmd) prepOtherRepos() {
	cl := cli.Client()
	maxRepoPrep := 50
	repos, err := cl.Repos.List(cli.Ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: sourcegraph.ListOptions{PerPage: int32(maxRepoPrep)},
	})
	if err != nil {
		log.Printf("Failed to prep other repos for build %v, go get for deps may fail", *c)
		return
	}

	if len(repos.Repos) == maxRepoPrep {
		log.Printf("Too many other repos to checkout for build %v, go get for deps may fail", *c)
		return
	}

	par := parallel.NewRun(4)
	for _, repo := range repos.Repos {
		repo := repo
		// Normal prep will take care of c.Repo
		if repo.URI == c.Repo {
			continue
		}
		// We only need to clone repos which require auth
		if !(repo.Origin == "" && !repo.Mirror) && !(repo.Private && repo.Mirror) {
			continue
		}
		par.Do(func() error {
			cmd := &prepBuildCmd{
				Repo:      repo.URI,
				BuildDir:  filepath.Join(c.PrivateWorkspace, "src", repo.URI),
				forcePrep: true,
			}
			err := cmd.Execute(nil)
			if err != nil {
				log.Printf("Failed to checkout potential dep %v for build %v, go get for deps may fail: %s", repo.URI, *c, err)
			}
			return nil
		})
	}
	par.Wait()

	gopath := c.PrivateWorkspace
	if curGopath := os.Getenv("GOPATH"); curGopath != "" {
		gopath = curGopath + string(filepath.ListSeparator) + gopath
	}
	os.Setenv("GOPATH", gopath)
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
