package fs

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/net/context"

	"strconv"
	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/store"
)

// BuildLogs is a local FS-backed implementation of the BuildLogs
// store.
//
// TODO(sqs): use the same dir as the other services? right now this
// uses conf.BuildLogDir, which is weird and inconsistent.
type BuildLogs struct{}

var _ store.BuildLogs = (*BuildLogs)(nil)

func (s *BuildLogs) Get(ctx context.Context, build sourcegraph.BuildSpec, tag, minIDStr string, minTime, maxTime time.Time) (*sourcegraph.LogEntries, error) {
	logPath := tag + ".log"

	// Clean the path of any relative parts so that we refuse to read files
	// outside the build log dir.
	if !strings.HasPrefix(logPath, "/") {
		logPath = "/" + logPath
	}
	logPath = path.Clean(logPath)[1:]

	// Read the log file.
	f := filepath.Join(conf.BuildLogDir, logPath)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		if os.IsNotExist(err) {
			return &sourcegraph.LogEntries{}, nil
		}
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	minID, err := strconv.Atoi(minIDStr)
	if err != nil && minIDStr != "" {
		return nil, err
	}
	if minID > len(lines) {
		minID = len(lines)
	}
	return &sourcegraph.LogEntries{Entries: lines[minID:], MaxID: strconv.Itoa(len(lines))}, nil
}
