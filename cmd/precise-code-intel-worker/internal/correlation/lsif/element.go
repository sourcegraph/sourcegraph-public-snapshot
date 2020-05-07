package lsif

import (
	"encoding/json"
	"strconv"
)

type ID string

type Element struct {
	ID    string
	Type  string
	Label string
	Raw   json.RawMessage
}

func UnmarshalElement(raw []byte) (Element, error) {
	var payload struct {
		ID    ID     `json:"id"`
		Type  string `json:"type"`
		Label string `json:"label"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Element{}, err
	}

	return Element{
		ID:    string(payload.ID),
		Type:  payload.Type,
		Label: payload.Label,
		Raw:   json.RawMessage(raw),
	}, nil
}

type Edge struct {
	OutV     string
	InV      string
	InVs     []string
	Document string
}

func UnmarshalEdge(element Element) (Edge, error) {
	var payload struct {
		OutV     ID   `json:"outV"`
		InV      ID   `json:"inV"`
		InVs     []ID `json:"inVs"`
		Document ID   `json:"document"`
	}
	if err := json.Unmarshal(element.Raw, &payload); err != nil {
		return Edge{}, err
	}

	var inVs []string
	for _, inV := range payload.InVs {
		inVs = append(inVs, string(inV))
	}

	return Edge{
		OutV:     string(payload.OutV),
		InV:      string(payload.InV),
		InVs:     inVs,
		Document: string(payload.Document),
	}, nil
}

func (id *ID) UnmarshalJSON(raw []byte) error {
	if raw[0] == '"' {
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}

		*id = ID(v)
		return nil
	}

	var v int64
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}

	*id = ID(strconv.FormatInt(v, 10))
	return nil
}
