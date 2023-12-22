package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func cloneStatus(cloned, cloning bool) types.CloneStatus {
	switch {
	case cloned:
		return types.CloneStatusCloned
	case cloning:
		return types.CloneStatusCloning
	}
	return types.CloneStatusNotCloned
}

func isAlwaysCloningTest(name api.RepoName) bool {
	return protocol.NormalizeRepo(name).Equal("github.com/sourcegraphtest/alwayscloningtest")
}

// repoLastFetched returns the mtime of the repo's FETCH_HEAD, which is the date of the last successful `git remote
// update` or `git fetch` (even if nothing new was fetched). As a special case when the repo has been cloned but
// none of those other two operations have been run (and so FETCH_HEAD does not exist), it will return the mtime of HEAD.
//
// This breaks on file systems that do not record mtime and if Git ever changes this undocumented behavior.
var repoLastFetched = func(dir common.GitDir) (time.Time, error) {
	fi, err := os.Stat(dir.Path("FETCH_HEAD"))
	if os.IsNotExist(err) {
		fi, err = os.Stat(dir.Path("HEAD"))
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// repoLastChanged returns the mtime of the repo's sg_refhash, which is the
// cached timestamp of the most recent commit we could find in the tree. As a
// special case when sg_refhash is missing we return repoLastFetched(dir).
//
// This breaks on file systems that do not record mtime. This is a Sourcegraph
// extension to track last time a repo changed. The file is updated by
// setLastChanged via doBackgroundRepoUpdate.
//
// As a special case, tries both the directory given, and the .git subdirectory,
// because we're a bit inconsistent about which name to use.
var repoLastChanged = func(dir common.GitDir) (time.Time, error) {
	fi, err := os.Stat(dir.Path("sg_refhash"))
	if os.IsNotExist(err) {
		return repoLastFetched(dir)
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// writeCounter wraps an io.Writer and keeps track of bytes written.
type writeCounter struct {
	w io.Writer
	// n is the number of bytes written to w
	n int64
}

func (c *writeCounter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.n += int64(n)
	return
}

// limitWriter is a io.Writer that writes to an W but discards after N bytes.
type limitWriter struct {
	W io.Writer // underling writer
	N int       // max bytes remaining
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if l.N <= 0 {
		return len(p), nil
	}
	origLen := len(p)
	if len(p) > l.N {
		p = p[:l.N]
	}
	n, err := l.W.Write(p)
	l.N -= n
	if l.N <= 0 {
		// If we have written limit bytes, then we can include the discarded
		// part of p in the count.
		n = origLen
	}
	return n, err
}

// mapToLoggerField translates a map to log context fields.
func mapToLoggerField(m map[string]any) []log.Field {
	LogFields := []log.Field{}

	for i, v := range m {

		LogFields = append(LogFields, log.String(i, fmt.Sprint(v)))
	}

	return LogFields
}

// hostnameMatch checks whether the hostname matches the given address.
// If we don't find an exact match, we look at the initial prefix.
func hostnameMatch(shardID, addr string) bool {
	if !strings.HasPrefix(addr, shardID) {
		return false
	}
	if addr == shardID {
		return true
	}
	// We know that shardID is shorter than addr so we can safely check the next
	// char
	next := addr[len(shardID)]
	return next == '.' || next == ':'
}

// Send 1 in 16 events to honeycomb. This is hardcoded since we only use this
// for Sourcegraph.com.
//
// 2020-05-29 1 in 4. We are currently at the top tier for honeycomb (before
// enterprise) and using double our quota. This gives us room to grow. If you
// find we keep bumping this / missing data we care about we can look into
// more dynamic ways to sample in our application code.
//
// 2020-07-20 1 in 16. Again hitting very high usage. Likely due to recent
// scaling up of the indexed search cluster. Will require more investigation,
// but we should probably segment user request path traffic vs internal batch
// traffic.
//
// 2020-11-02 Dynamically sample. Again hitting very high usage. Same root
// cause as before, scaling out indexed search cluster. We update our sampling
// to instead be dynamic, since "rev-parse" is 12 times more likely than the
// next most common command.
//
// 2021-08-20 over two hours we did 128 * 128 * 1e6 rev-parse requests
// internally. So we update our sampling to heavily downsample internal
// rev-parse, while upping our sampling for non-internal.
// https://ui.honeycomb.io/sourcegraph/datasets/gitserver-exec/result/67e4bLvUddg
func honeySampleRate(cmd string, actor *actor.Actor) uint {
	// HACK(keegan) 2022-11-02 IsInternal on sourcegraph.com is always
	// returning false. For now I am also marking it internal if UID is not
	// set to work around us hammering honeycomb.
	internal := actor.IsInternal() || actor.UID == 0
	switch {
	case cmd == "rev-parse" && internal:
		return 1 << 14 // 16384

	case internal:
		// we care more about user requests, so downsample internal more.
		return 16

	default:
		return 8
	}
}

// shortGitCommandSlow returns the threshold for regarding an git command as
// slow. Some commands such as "git archive" are inherently slower than "git
// rev-parse", so this will return an appropriate threshold given the command.
func shortGitCommandSlow(args []string) time.Duration {
	if len(args) < 1 {
		return time.Second
	}
	switch args[0] {
	case "archive":
		return 1 * time.Minute

	case "blame", "ls-tree", "log", "show":
		return 5 * time.Second

	default:
		return 2500 * time.Millisecond
	}
}
