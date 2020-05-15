package lsif

import (
	"encoding/json"
)

type PackageInformationData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func UnmarshalPackageInformationData(element Element) (payload PackageInformationData, err error) {
	err = json.Unmarshal(element.Raw, &payload)
	return payload, err
}
