package lsif

import (
	"encoding/json"
)

type Element struct {
	ID    string `json:"id"` // TODO - string or int
	Type  string `json:"type"`
	Label string `json:"label"`
	Raw   json.RawMessage
}

func UnmarshalElement(Raw []byte) (payload Element, err error) {
	err = json.Unmarshal(Raw, &payload)
	payload.Raw = json.RawMessage(Raw)
	return payload, err
}

type Edge struct {
	OutV     string   `json:"outV"`
	InV      string   `json:"inV"`
	InVs     []string `json:"inVs"`
	Document string   `json:"document"`
}

func UnmarshalEdge(element Element) (payload Edge, err error) {
	err = json.Unmarshal(element.Raw, &payload)
	return payload, err
}
