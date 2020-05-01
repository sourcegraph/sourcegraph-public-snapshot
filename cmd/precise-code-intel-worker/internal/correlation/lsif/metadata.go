package lsif

import (
	"encoding/json"
	"strings"
)

type MetaData struct {
	Version     string `json:"version"`
	ProjectRoot string `json:"projectRoot"`
}

func UnmarshalMetaData(element Element, dumpRoot string) (payload MetaData, err error) {
	err = json.Unmarshal(element.Raw, &payload)

	// We assume that the project root in the LSIF dump is either:
	//
	//   (1) the root of the LSIF dump, or
	//   (2) the root of the repository
	//
	// These are the common cases and we don't explicitly support
	// anything else. Here we normalize to (1) by appending the dump
	// root if it's not already suffixed by it.

	if !strings.HasSuffix(payload.ProjectRoot, "/") {
		payload.ProjectRoot += "/"
	}

	if dumpRoot != "" && !strings.HasPrefix(payload.ProjectRoot, dumpRoot) {
		payload.ProjectRoot += dumpRoot
	}

	return payload, err
}
