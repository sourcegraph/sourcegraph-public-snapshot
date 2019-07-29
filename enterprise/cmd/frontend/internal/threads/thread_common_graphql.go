package threads

import (
	"context"
	"path"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// GQLThreadCommon implements the GraphQL type Thread.
type GQLThreadCommon struct{ db *dbThread }

func (v *GQLThreadCommon) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.db.RepositoryID)
}

func (v *GQLThreadCommon) Number() string { return strconv.FormatInt(v.db.ID, 10) }

func (v *GQLThreadCommon) DBID() int64 { return v.db.ID }

func (v *GQLThreadCommon) Title() string { return v.db.Title }

func (v *GQLThreadCommon) ExternalURL() *string { return v.db.ExternalURL }

func (v *GQLThreadCommon) URL(ctx context.Context) (string, error) {
	repository, err := v.Repository(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(repository.URL(), "-", "threads", v.Number()), nil
}
