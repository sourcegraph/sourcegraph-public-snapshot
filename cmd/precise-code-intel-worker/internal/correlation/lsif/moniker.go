package lsif

import (
	"encoding/json"
)

type MonikerData struct {
	Kind                 string `json:"kind"`
	Scheme               string `json:"scheme"`
	Identifier           string `json:"identifier"`
	PackageInformationID string
}

func UnmarshalMonikerData(element Element) (payload MonikerData, err error) {
	err = json.Unmarshal(element.Raw, &payload)
	if payload.Kind == "" {
		payload.Kind = "local"
	}
	return payload, err
}

func (d MonikerData) SetPackageInformationID(id string) MonikerData {
	return MonikerData{
		Kind:                 d.Kind,
		Scheme:               d.Scheme,
		Identifier:           d.Identifier,
		PackageInformationID: id,
	}
}
