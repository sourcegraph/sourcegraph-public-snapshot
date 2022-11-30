package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type RepoMetadata struct {
	ID        api.RepoID
	CreatedAt time.Time
	UpdatedAt time.Time
	Ignored   bool
}

func (meta *RepoMetadata) Cursor() int64 { return int64(meta.ID) }
