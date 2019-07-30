package threadlike

import (
	"context"
	"path"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// GQLThreadlike implements common fields for the GraphQL thread, issue, and changeset types.
type GQLThreadlike struct{ DB *internal.DBThread }

func (v *GQLThreadlike) Type() graphqlbackend.ThreadlikeType { return v.DB.Type }

func (v *GQLThreadlike) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.DB.RepositoryID)
}

func (v *GQLThreadlike) Number() string { return strconv.FormatInt(v.DB.ID, 10) }

func (v *GQLThreadlike) DBID() int64 { return v.DB.ID }

func (v *GQLThreadlike) Title() string { return v.DB.Title }

func (v *GQLThreadlike) Body() *string { return nil /* TODO!(sqs) */ }

func (v *GQLThreadlike) ExternalURL() *string { return v.DB.ExternalURL }

func (v *GQLThreadlike) URL(ctx context.Context) (string, error) {
	repository, err := v.Repository(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(repository.URL(), "-", "threads", v.Number()), nil
}
