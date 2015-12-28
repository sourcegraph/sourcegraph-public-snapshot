package fs

import (
	"io/ioutil"
	"testing"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func writeBuildLog(ctx context.Context, t *testing.T, task sourcegraph.TaskSpec, data string) {
	if err := ioutil.WriteFile(logFilePath(task), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func init() {
	tmpDir, err := ioutil.TempDir("", "BuildLogs-test")
	if err != nil {
		panic(err)
	}
	conf.BuildLogDir = tmpDir
}

func TestBuildLogs_Get_noErrorIfNotExist(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.BuildLogs_Get_noErrorIfNotExist(ctx, t, &BuildLogs{}, writeBuildLog)
}

func TestBuildLogs_Get_noErrorIfEmpty(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.BuildLogs_Get_noErrorIfEmpty(ctx, t, &BuildLogs{}, writeBuildLog)
}

func TestBuildLogs_Get_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.BuildLogs_Get_ok(ctx, t, &BuildLogs{}, writeBuildLog)
}

func TestBuildLogs_Get_MinID(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.BuildLogs_Get_MinID(ctx, t, &BuildLogs{}, writeBuildLog)
}
