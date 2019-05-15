package graphqlbackend

import (
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
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
