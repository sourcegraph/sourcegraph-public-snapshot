package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

// clearAllIndexes clears all indexes from the target instance.
func clearAllIndexes(ctx context.Context) error {
	client := internal.GraphQLClient()

	for {
		if requery, err := clearIndexesOnce(ctx, client); err != nil {
			return err
		} else if !requery {
			break
		}

		<-time.After(time.Second)
	}

	fmt.Printf("[%5s] %s All indexes deleted\n", internal.TimeSince(start), internal.EmojiSuccess)
	return nil
}

func clearIndexesOnce(_ context.Context, client *gqltestutil.Client) (requery bool, _ error) {
	var payload struct {
		Data struct {
			LSIFIndexes struct {
				Nodes []jsonIndexResult
			}
		}
	}
	if err := client.GraphQL(internal.SourcegraphAccessToken, indexesQuery, nil, &payload); err != nil {
		return false, err
	}

	for _, index := range payload.Data.LSIFIndexes.Nodes {
		// TODO - display repo@commit instead
		fmt.Printf("[%5s] %s Deleting index %s\n", internal.TimeSince(start), internal.EmojiLightbulb, index.ID)

		if err := client.GraphQL(internal.SourcegraphAccessToken, deleteIndexQuery, map[string]any{"id": index.ID}, nil); err != nil {
			return false, err
		}

		requery = true
	}

	return requery, nil
}

type jsonIndexResult struct {
	ID    string
	State string
}

const indexesQuery = `
query CodeIntelQA_Clear_Indexes {
	lsifIndexes {
		nodes {
			id
			state
		}
	}
}
`

const deleteIndexQuery = `
mutation CodeIntelQA_Clear_DeleteIndex($id: ID!) {
	deleteLSIFIndex(id: $id) {
		alwaysNil
	}
}
`

// clearAllUploads clears all uploads from the target instance.
func clearAllUploads(ctx context.Context) error {
	client := internal.GraphQLClient()

	for {
		if requery, err := clearUploadsOnce(ctx, client); err != nil {
			return err
		} else if !requery {
			break
		}

		<-time.After(time.Second)
	}

	fmt.Printf("[%5s] %s All uploads deleted\n", internal.TimeSince(start), internal.EmojiSuccess)
	return nil
}

func clearUploadsOnce(_ context.Context, client *gqltestutil.Client) (requery bool, _ error) {
	var payload struct {
		Data struct {
			LSIFUploads struct {
				Nodes []jsonUploadResult
			}
		}
	}
	if err := client.GraphQL(internal.SourcegraphAccessToken, uploadsQuery, nil, &payload); err != nil {
		return false, err
	}

	purging := make([]jsonUploadResult, 0, len(payload.Data.LSIFUploads.Nodes))
	for _, upload := range payload.Data.LSIFUploads.Nodes {
		if upload.State == "DELETED" {
			continue
		}

		if upload.State == "DELETING" {
			purging = append(purging, upload)
		} else {
			// TODO - display repo@commit instead
			fmt.Printf("[%5s] %s Deleting upload %s\n", internal.TimeSince(start), internal.EmojiLightbulb, upload.ID)

			if err := client.GraphQL(internal.SourcegraphAccessToken, deleteUploadQuery, map[string]any{"id": upload.ID}, nil); err != nil {
				return false, err
			}
		}

		requery = true
	}

	if !requery && len(purging) > 0 {
		for _, upload := range purging {
			// TODO - display repo@commit instead
			fmt.Printf("[%5s] %s Waiting for upload %s to be purged\n", internal.TimeSince(start), internal.EmojiLightbulb, upload.ID)

		}

		requery = true
	}

	return requery, nil
}

type jsonUploadResult struct {
	ID    string
	State string
}

const uploadsQuery = `
query CodeIntelQA_Clear_Uploads {
	lsifUploads {
		nodes {
			id
			state
		}
	}
}
`

const deleteUploadQuery = `
mutation CodeIntelQA_Clear_DeleteUploads($id: ID!) {
	deleteLSIFUpload(id: $id) {
		alwaysNil
	}
}
`
