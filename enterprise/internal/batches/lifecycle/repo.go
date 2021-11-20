package lifecycle

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type repoMarshaller struct {
	repo *types.Repo
}

func (rm repoMarshaller) MarshalJSON() ([]byte, error) {
	// TODO: figure out what fields should be excluded here.
	return json.Marshal(rm.repo)
}
