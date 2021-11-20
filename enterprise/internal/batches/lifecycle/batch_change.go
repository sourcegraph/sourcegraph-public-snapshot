package lifecycle

import (
	"encoding/json"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

type batchChangeMarshaller struct {
	batchChange *btypes.BatchChange
}

func (bcm *batchChangeMarshaller) MarshalJSON() ([]byte, error) {
	// TODO: figure out what fields should be excluded.
	return json.Marshal(bcm.batchChange)
}
