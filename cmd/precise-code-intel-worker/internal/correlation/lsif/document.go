package lsif

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
)

type DocumentData struct {
	URI      string `json:"uri"`
	Contains datastructures.IDSet
}

func UnmarshalDocumentData(element Element, projectRoot string) (payload DocumentData, err error) {
	err = json.Unmarshal(element.Raw, &payload)
	if !strings.HasPrefix(payload.URI, projectRoot) {
		return DocumentData{}, fmt.Errorf("document URI %s is not relative to project root %s", payload.URI, projectRoot)
	}
	payload.URI = payload.URI[len(projectRoot):]
	payload.Contains = datastructures.IDSet{}
	return payload, err
}
