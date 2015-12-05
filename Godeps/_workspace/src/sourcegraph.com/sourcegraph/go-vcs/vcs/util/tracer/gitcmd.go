package tracer

import (
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
)

// gitcmdCrossRepo wraps a gitcmd.CrossRepo, adding tracing to it.
type gitcmdCrossRepo struct {
	c   gitcmd.CrossRepo
	rec *appdash.Recorder
}

// GitRootDir implements the gitcmd.CrossRepo interface.
func (c gitcmdCrossRepo) GitRootDir() string {
	start := time.Now()
	dir := c.c.GitRootDir()
	c.rec.Child().Event(GoVCS{
		Name:      "gitcmd.CrossRepo.GitRootDir",
		StartTime: start,
		EndTime:   time.Now(),
	})
	return dir
}
