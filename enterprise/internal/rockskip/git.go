package rockskip

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type Git interface {
	LogReverseEach(repo string, db database.DB, commit string, n int, onLogEntry func(logEntry gitdomain.LogEntry) error) error
	RevListEach(repo string, db database.DB, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error
	ArchiveEach(repo string, commit string, paths []string, onFile func(path string, contents []byte) error) error
}
