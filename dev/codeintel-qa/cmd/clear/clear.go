package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

// clearAllPreciseIndexes clears all precise indexes from the target instance.
func clearAllPreciseIndexes(ctx context.Context) error {
	client := internal.GraphQLClient()

	for {
		if requery, err := clearPreciseIndexesOnce(ctx, client); err != nil {
			return err
		} else if !requery {
			break
		}

		<-time.After(time.Second)
	}

	fmt.Printf("[%5s] %s All precise indexes deleted\n", internal.TimeSince(start), internal.EmojiSuccess)
	return nil
}

func clearPreciseIndexesOnce(_ context.Context, client *gqltestutil.Client) (requery bool, _ error) {
	var payload struct {
		Data struct {
			PreciseIndexes struct {
				Nodes []jsonPreciseIndexResult
			}
		}
	}
	if err := client.GraphQL(internal.SourcegraphAccessToken, precisesIndexesQuery, nil, &payload); err != nil {
		return false, err
	}

	purging := make([]jsonPreciseIndexResult, 0, len(payload.Data.PreciseIndexes.Nodes))
	for _, preciseIndex := range payload.Data.PreciseIndexes.Nodes {
		if preciseIndex.State == "DELETED" {
			continue
		}

		if preciseIndex.State == "DELETING" {
			purging = append(purging, preciseIndex)
		} else {
			// TODO - display repo@commit instead
			fmt.Printf("[%5s] %s Deleting precise index %s\n", internal.TimeSince(start), internal.EmojiLightbulb, preciseIndex.ID)

			if err := client.GraphQL(internal.SourcegraphAccessToken, deletePreciseIndexQuery, map[string]any{"id": preciseIndex.ID}, nil); err != nil {
				return false, err
			}
		}

		requery = true
	}

	if !requery && len(purging) > 0 {
		for _, preciseIndex := range purging {
			// TODO - display repo@commit instead
			fmt.Printf("[%5s] %s Waiting for precise index %s to be purged\n", internal.TimeSince(start), internal.EmojiLightbulb, preciseIndex.ID)

		}

		requery = true
	}

	return requery, nil
}

type jsonPreciseIndexResult struct {
	ID    string
	State string
}

const precisesIndexesQuery = `
query CodeIntelQA_Clear_PreciseIndexes {
	preciseIndexes {
		nodes {
			id
			state
		}
	}
}
`

const deletePreciseIndexQuery = `
mutation CodeIntelQA_Clear_DeletePreciseIndex($id: ID!) {
	deletePreciseIndex(id: $id) {
		alwaysNil
	}
}
`
