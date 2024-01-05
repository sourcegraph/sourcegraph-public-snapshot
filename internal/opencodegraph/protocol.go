package opencodegraph

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/schema"
)

func DecodeRequestMessage(d *json.Decoder) (method string, cap *schema.CapabilitiesParams, ann *schema.AnnotationsParams, err error) {
	var req struct {
		schema.RequestMessage
		Params json.RawMessage `json:"params"`
	}
	if err := d.Decode(&req); err != nil {
		return "", nil, nil, err
	}

	method = req.Method
	switch method {
	case "capabilities":
		err = json.Unmarshal(req.Params, &cap)
	case "annotations":
		err = json.Unmarshal(req.Params, &ann)
	}

	return
}
