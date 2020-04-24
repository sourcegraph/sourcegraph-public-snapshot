package server

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type metaDataVertex struct {
	Label    string   `json:"label"`
	ToolInfo toolInfo `json:"toolInfo"`
}

type toolInfo struct {
	Name string `json:"name"`
}

var ErrInvalidMetaDataVertex = errors.New("invalid metadata vertex")

// readIndexerNameFromFile returns the name of the tool that generated
// the given index file. This function reads only the first line of the
// file, where the metadata vertex is assumed to be in all valid dumps.
// This function also resets the offset of the file to the beginning of
// the file before and after reading.
func readIndexerNameFromFile(f *os.File) (string, error) {
	_, err1 := f.Seek(0, 0)
	name, err2 := readIndexerName(f)
	_, err3 := f.Seek(0, 0)

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			return "", err
		}
	}

	return name, nil
}

// readIndexerName returns the name of the tool that generated the given
// index contents. This function reads only the first line of the file,
// where the metadata vertex is assumed to be in all valid dumps.
func readIndexerName(r io.Reader) (string, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}

	line, isPrefix, err := bufio.NewReader(gzipReader).ReadLine()
	if err != nil {
		return "", err
	}
	if isPrefix {
		return "", errors.New("metaData vertex exceeds buffer")
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
