package shared

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

func downloadJSONFile[T any](ctx context.Context, uploadStore uploadstore.Store, key string) (*T, error) {
	file, err := uploadStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var jsonFile T
	err = json.NewDecoder(file).Decode(&jsonFile)
	if err != nil {
		return nil, err
	}
	return &jsonFile, nil
}
