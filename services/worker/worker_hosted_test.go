// +build exectest,buildtest

package worker_test

import (
	"strings"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

func TestWorker_BuildRepo_noSrclib_pass(t *testing.T) {
	_, done, build, buildLog := testWorker_buildRepo(t, map[string]string{
		".drone.yml": `
build:
  image: library/alpine:3.2
  commands:
    - echo PASS 12345
`,
	})
	defer done()

	if !build.Success {
		t.Log(buildLog)
		t.Errorf("build %s failed (want it to succeed)", build.Spec().IDString())
	}
	if want := "PASS 12345"; !strings.Contains(buildLog, want) {
		t.Errorf("build log doesn't contain %q\n\n%s", want, buildLog)
	}
}

func TestWorker_BuildRepo_noSrclib_fail(t *testing.T) {
	_, done, build, buildLog := testWorker_buildRepo(t, map[string]string{
		".drone.yml": `
build:
  image: library/alpine:3.2
  commands:
    - echo FAIL 12345
    - exit 1
`,
	})
	defer done()

	if build.Success {
		t.Fatalf("build %s succeeded (want it to fail)", build.Spec().IDString())
	}
	if want := "FAIL 12345"; !strings.Contains(buildLog, want) {
		t.Errorf("build log doesn't contain %q\n\n%s", want, buildLog)
	}
}

// Test when the repo doesn't contain any srclib auto-detected
// languages, but it does explicitly configure srclib analysis.
func TestWorker_BuildRepo_srclibExplicit_pass(t *testing.T) {
	t.Skip("flaky") // https://app.asana.com/0/87040567695724/93915183753618

	_, _, sampleImage := testserver.SrclibSampleToolchain(true)

	ctx, done, build, buildLog := testWorker_buildRepo(t, map[string]string{
		"f": "f",
		".drone.yml": `
build:
  srclib-sample:
    image: ` + sampleImage + `
    commands:
      - srclib config
      - srclib make
`,
	})
	defer done()

	if !build.Success {
		t.Log(buildLog)
		t.Fatalf("build %s failed (want it to succeed)", build.Spec().IDString())
	}
	if want := "Importing to "; !strings.Contains(buildLog, want) {
		t.Errorf("build log doesn't contain %q\n\n%s", want, buildLog)
	}

	testutil.CheckImport(t, ctx, build.Repo, build.CommitID)
}

func testWorker_buildRepo(t *testing.T, files map[string]string) (ctx context.Context, done func(), build *sourcegraph.Build, buildLog string) {
	t.Parallel()

	a, ctx := testserver.NewServer()

	// Create and push a repo that uses the sample toolchain.
	repo, _, repoDone, err := testutil.CreateAndPushRepoFiles(t, ctx, "r/r", files)
	if err != nil {
		t.Fatal(err)
	}

	done = func() {
		repoDone()
		a.Close()
	}

	buildSpec := sourcegraph.BuildSpec{Repo: repo.RepoSpec(), ID: 1}

	// Get log for a single task.
	getTaskLog := func(task sourcegraph.TaskSpec) (string, error) {
		log, err := a.Client.Builds.GetTaskLog(ctx, &sourcegraph.BuildsGetTaskLogOp{
			Task: task,
			Opt:  &sourcegraph.BuildGetLogOptions{},
		})
		if err != nil {
			return "", err
		}
		return strings.Join(log.Entries, "\n"), nil
	}

	// Get the combined log for all tasks in the build.
	getLog := func() (string, error) {
		tasks, err := a.Client.Builds.ListBuildTasks(ctx, &sourcegraph.BuildsListBuildTasksOp{
			Build: buildSpec,
			Opt:   &sourcegraph.BuildTaskListOptions{ListOptions: sourcegraph.ListOptions{PerPage: 1000}},
		})
		if err != nil {
			return "", err
		}

		var logs []string
		for _, task := range tasks.BuildTasks {
			log, err := getTaskLog(task.Spec())
			if err != nil {
				return "", err
			}
			logs = append(logs, log)
		}
		return strings.Join(logs, "\n"), nil
	}

	// Pushing triggers a build; wait for it to finish.
	build, err = testutil.WaitForBuild(t, ctx, buildSpec)
	if err != nil {
		if build != nil {
			if log, err := getLog(); err == nil {
				t.Logf("log:\n%s", log)
			}
		}
		t.Fatal(err)
	}

	buildLog, err = getLog()
	if err != nil {
		t.Fatal(err)
	}

	return ctx, done, build, buildLog
}
