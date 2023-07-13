package graphql

import (
	"errors"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

func validateGetPreciseContextInput(input *resolverstubs.GetPreciseContextInput) error {
	if input.Input.ActiveFile == "" {
		return errors.New("active file must be set")
	}
	if input.Input.ActiveFileContent == "" {
		return errors.New("active file content must be set")
	}
	if input.Input.CommitID == "" {
		return errors.New("commit ID must be set")
	}

	return nil
}
