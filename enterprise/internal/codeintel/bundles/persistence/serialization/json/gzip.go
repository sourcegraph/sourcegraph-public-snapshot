package json

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

// TODO(efritz) - document
func compress(uncompressed []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := io.Copy(gzipWriter, bytes.NewReader(uncompressed)); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// unmarshalGzippedJSON unmarshals the gzip+json encoded data.
func unmarshalGzippedJSON(data []byte, payload interface{}) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}

	return json.NewDecoder(gzipReader).Decode(&payload)
}
