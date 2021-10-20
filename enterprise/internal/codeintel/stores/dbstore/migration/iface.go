package migration

import (
	"context"
	"time"
)

type GitserverClient interface {
	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
}
