package upload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MaxBufferSize is the maximum size of the metaData line in the dump. This should be large enough
// to be able to read the output of lsif-tsc for most cases, which will contain all glob-expanded
// file names in the indexing of JavaScript projects.
//
// Data point: lodash's metaData vertex constructed by the args `*.js test/*.js --AllowJs --checkJs`
// is 10639 characters long.
const MaxBufferSize = 128 * 1024

// ErrMetadataExceedsBuffer occurs when the first line of an LSIF index is too long to read.
var ErrMetadataExceedsBuffer = errors.New("metaData vertex exceeds buffer")

// ErrInvalidMetaDataVertex occurs when the first line of an LSIF index is not a valid metadata vertex.
var ErrInvalidMetaDataVertex = errors.New("invalid metaData vertex")

type metaDataVertex struct {
	Label    string   `json:"label"`
	ToolInfo toolInfo `json:"toolInfo"`
}

type toolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ReadIndexerName returns the name of the tool that generated the given index contents.
// This function reads only the first line of the file, where the metadata vertex is
// assumed to be in all valid dumps.
func ReadIndexerName(r io.Reader) (string, error) {
	name, _, err := ReadIndexerNameAndVersion(r)
	return name, err
}

// ReadIndexerNameAndVersion returns the name and version of the tool that generated the
// given index contents. This function reads only the first line of the file for LSIF, where
// the metadata vertex is assumed to be in all valid dumps. If its a SCIP index, the name
// and version are read from the contents of the index.
func ReadIndexerNameAndVersion(r io.Reader) (name string, verison string, _ error) {
	var buf bytes.Buffer
	line, isPrefix, err := bufio.NewReaderSize(io.TeeReader(r, &buf), MaxBufferSize).ReadLine()
	if err == nil {
		if !isPrefix {
			meta := metaDataVertex{}
			if err := json.Unmarshal(line, &meta); err == nil {
				if meta.Label == "metaData" && meta.ToolInfo.Name != "" {
					return meta.ToolInfo.Name, meta.ToolInfo.Version, nil
				}
			}
		}
	}

	content, err := io.ReadAll(io.MultiReader(bytes.NewReader(buf.Bytes()), r))
	if err != nil {
		return "", "", ErrInvalidMetaDataVertex
	}

	var index scip.Index
	if err := proto.Unmarshal(content, &index); err != nil {
		return "", "", ErrInvalidMetaDataVertex
	}

	return index.Metadata.ToolInfo.Name, index.Metadata.ToolInfo.Version, nil
}
