package lsif

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
)

type DocumentData struct {
	URI      string `json:"uri"`
	Contains datastructures.IDSet
}

func UnmarshalDocumentData(element Element, projectRoot string) (payload DocumentData, err error) {
	err = json.Unmarshal(element.Raw, &payload)
	if err != nil {
		return DocumentData{}, err
	}

	relativeURI, err := filepath.Rel(projectRoot, payload.URI)
	if err != nil {
		return DocumentData{}, fmt.Errorf("document URI %q is not relative to project root %q (%s)", payload.URI, projectRoot, err)
	}

	payload.URI = relativeURI
	payload.Contains = datastructures.IDSet{}
	return payload, err
}
