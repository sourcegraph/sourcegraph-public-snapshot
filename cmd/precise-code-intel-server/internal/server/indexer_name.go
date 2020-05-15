package server

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

type metaDataVertex struct {
	Label    string   `json:"label"`
	ToolInfo toolInfo `json:"toolInfo"`
}

type toolInfo struct {
	Name string `json:"name"`
}

// ErrMetadataExceedsBuffer occurs when the first line of an LSIF index is too long to read.
var ErrMetadataExceedsBuffer = errors.New("metaData vertex exceeds buffer")

// ErrInvalidMetaDataVertex occurs when the first line of an LSIF index is not a valid metadata vertex.
var ErrInvalidMetaDataVertex = errors.New("invalid metaData vertex")

// readIndexerName returns the name of the tool that generated the given index contents.
// This function reads only the first line of the file, where the metadata vertex is
// assumed to be in all valid dumps.
func readIndexerName(r io.Reader) (string, error) {
	line, isPrefix, err := bufio.NewReader(r).ReadLine()
	if err != nil {
		return "", err
	}
	if isPrefix {
		return "", ErrMetadataExceedsBuffer
	}

	meta := metaDataVertex{}
	if err := json.Unmarshal(line, &meta); err != nil {
		return "", ErrInvalidMetaDataVertex
	}

	if meta.Label != "metaData" || meta.ToolInfo.Name == "" {
		return "", ErrInvalidMetaDataVertex
	}

	return meta.ToolInfo.Name, nil
}
