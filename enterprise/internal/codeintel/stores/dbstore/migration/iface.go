package migration

import (
	"context"
	"time"
)

type GitserverClient interface {
	CommitDate(ctx context.Context, repositoryID int, commit string) (time.Time, error)
}
