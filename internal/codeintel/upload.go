package codeintel

import (
	"encoding/base64"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/utils"
)

type UploadIndexOpts = codeintelutils.UploadIndexOpts

func UploadIndex(opts UploadIndexOpts) (string, error) {
	id, err := codeintelutils.UploadIndex(opts)
	if err != nil {
		return "", err
	}

	return uploadIDToGraphQLID(id), nil
}

// uploadIndex constructs a GraphQL-compatible identifier from the raw identifier returned
// from the upload endpoint.
func uploadIDToGraphQLID(uploadID int) string {
	return string(base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(`LSIFUpload:"%d"`, uploadID))))
}
