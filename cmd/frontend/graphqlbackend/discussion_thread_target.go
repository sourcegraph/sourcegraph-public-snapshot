package graphqlbackend

import (
	"context"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions"
)

func (r *discussionThreadTargetRepoResolver) ID() graphql.ID {
	return marshalDiscussionThreadTargetID(r.t.ID)
}

func marshalDiscussionThreadTargetID(dbID int64) graphql.ID {
	return relay.MarshalID("DiscussionThreadTarget", strconv.FormatInt(dbID, 36))
}

func unmarshalDiscussionThreadTargetID(id graphql.ID) (dbID int64, err error) {
	var dbIDStr string
	err = relay.UnmarshalSpec(id, &dbIDStr)
	if err == nil {
		dbID, err = strconv.ParseInt(dbIDStr, 36, 64)
	}
	return
}

func (r *discussionThreadTargetRepoResolver) IsIgnored() bool {
	return r.t.IsIgnored
}

func (r *discussionThreadTargetRepoResolver) URL(ctx context.Context) (*string, error) {
	// TODO!(sqs): Add threadID and commentID.
	url, err := discussions.URLToInlineTarget(ctx, r.t, nil, nil)
	if err != nil || url == nil {
		return nil, err
	}
	return strptr(url.String()), nil
}
