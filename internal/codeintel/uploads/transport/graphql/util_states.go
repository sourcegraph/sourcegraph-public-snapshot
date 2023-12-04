package graphql

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func bifurcateStates(states []string) (uploadStates, indexStates []string, _ error) {
	for _, state := range states {
		switch strings.ToUpper(state) {
		case "QUEUED_FOR_INDEXING":
			indexStates = append(indexStates, "queued")
		case "INDEXING":
			indexStates = append(indexStates, "processing")
		case "INDEXING_ERRORED":
			indexStates = append(indexStates, "errored")

		case "UPLOADING_INDEX":
			uploadStates = append(uploadStates, "uploading")
		case "QUEUED_FOR_PROCESSING":
			uploadStates = append(uploadStates, "queued")
		case "PROCESSING":
			uploadStates = append(uploadStates, "processing")
		case "PROCESSING_ERRORED":
			uploadStates = append(uploadStates, "errored")
		case "COMPLETED":
			uploadStates = append(uploadStates, "completed")
		case "DELETING":
			uploadStates = append(uploadStates, "deleting")
		case "DELETED":
			uploadStates = append(uploadStates, "deleted")

		default:
			return nil, nil, errors.Newf("filtering by state %q is unsupported", state)
		}
	}

	return uploadStates, indexStates, nil
}
