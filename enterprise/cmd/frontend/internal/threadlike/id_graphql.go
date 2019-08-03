package threadlike

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

type gqlType string

const (
	GQLTypeThread    gqlType = "Thread"
	GQLTypeIssue             = "Issue"
	GQLTypeChangeset         = "Changeset"
)

func MarshalID(typ gqlType, id int64) graphql.ID {
	return relay.MarshalID(string(typ), id)
}

func UnmarshalID(id graphql.ID) (typ gqlType, dbID int64, err error) {
	typ = gqlType(relay.UnmarshalKind(id))
	if typ != GQLTypeThread && typ != GQLTypeIssue && typ != GQLTypeChangeset {
		return "", 0, fmt.Errorf("invalid threadlike type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func UnmarshalIDOfType(typ gqlType, id graphql.ID) (dbID int64, err error) {
	gotType, dbID, err := UnmarshalID(id)
	if err != nil {
		return 0, err
	}
	if gotType != typ {
		// TODO!(sqs): uncomment, for demo
		//
		// return 0, fmt.Errorf("got threadlike ID of type %q, expected %q", gotType, typ)
	}
	return dbID, nil
}
