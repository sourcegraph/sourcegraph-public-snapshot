package graphql

import (
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func UnmarshalPreciseIndexGQLID(id graphql.ID) (uploadID, indexID int, err error) {
	uploadID, indexID, err = unmarshalRawPreciseIndexGQLID(id)
	if err == nil && uploadID == 0 && indexID == 0 {
		err = errors.Newf("invalid precise index id %q", id)
	}

	return uploadID, indexID, errors.Wrap(err, "unexpected precise index ID")
}

var errExpectedPairs = errors.New("expected pairs of `U:<id>`, `I:<id>`")

func unmarshalRawPreciseIndexGQLID(id graphql.ID) (uploadID, indexID int, err error) {
	rawPayload, err := resolverstubs.UnmarshalID[string](id)
	if err != nil {
		return 0, 0, errors.Wrap(err, "unexpected precise index ID")
	}

	parts := strings.Split(rawPayload, ":")
	if len(parts)%2 != 0 {
		return 0, 0, errExpectedPairs
	}
	for i := 0; i < len(parts)-1; i += 2 {
		id, err := strconv.Atoi(parts[i+1])
		if err != nil {
			return 0, 0, errExpectedPairs
		}

		switch parts[i] {
		case "U":
			uploadID = id
		case "I":
			indexID = id
		default:
			return 0, 0, errExpectedPairs
		}
	}

	return uploadID, indexID, nil
}
