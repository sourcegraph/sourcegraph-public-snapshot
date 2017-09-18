package graphqlbackend

import (
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgRepoResolver struct {
	org  *sourcegraph.Org
	repo *sourcegraph.OrgRepo
}

func (o *orgRepoResolver) ID() int32 {
	return o.repo.ID
}

func (o *orgRepoResolver) Org() *orgResolver {
	return &orgResolver{o.org}
}

func (o *orgRepoResolver) RemoteURI() string {
	return o.repo.RemoteURI
}

func (o *orgRepoResolver) CreatedAt() string {
	return o.repo.CreatedAt.Format(time.RFC3339)
}

func (o *orgRepoResolver) UpdatedAt() string {
	return o.repo.UpdatedAt.Format(time.RFC3339)
}
