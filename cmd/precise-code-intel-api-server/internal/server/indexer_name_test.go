package server

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const testMetaDataVertex = `{"label": "metaData", "toolInfo": {"name": "test"}}`
const testVertex = `{"id": "a", "type": "edge", "label": "textDocument/references", "outV": "b", "inV": "c"}`

func TestReadIndexerName(t *testing.T) {
	name, err := readIndexerName(generateTestIndex(testMetaDataVertex))
	if err != nil {
		t.Fatalf("unexpected error reading indexer name: %s", err)
	}
	if name != "test" {
		t.Errorf("unexpected indexer name. want=%s have=%s", "test", name)
	}
}

func TestReadIndexerNameMalformed(t *testing.T) {
	for _, metaDataVertex := range []string{`invalid json`, `{"label": "textDocument/references"}`} {
		if _, err := readIndexerName(generateTestIndex(metaDataVertex)); err != ErrInvalidMetaDataVertex {
			t.Fatalf("unexpected error reading indexer name. want=%q have=%q", ErrInvalidMetaDataVertex, err)
		}
	}
}

func TestReadIndexerNameFromFile(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp file: %s", err)
	}
	defer os.Remove(tempFile.Name())

	_, _ = io.Copy(tempFile, generateTestIndex(testMetaDataVertex))

	name, err := readIndexerNameFromFile(tempFile)
	if err != nil {
		t.Fatalf("unexpected error reading indexer name: %s", err)
	}
	if name != "test" {
		t.Errorf("unexpected indexer name. want=%s have=%s", "test", name)
	}

	// Ensure reader is reset to beginning
	firstLine, err := testReadFirstLine(tempFile)
	if err != nil {
		t.Fatalf("unexpected error reading from file %s", err)
	}
	if firstLine != testMetaDataVertex {
		t.Errorf("unexpected buffer location. want=%q have=%q", testMetaDataVertex, firstLine)
	}
}

func generateTestIndex(metaDataVertex string) io.Reader {
	lines := []string{metaDataVertex}
	for i := 0; i < 20000; i++ {
		lines = append(lines, testVertex)
	}

	content := strings.Join(lines, "\n") + "\n"

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = io.Copy(w, bytes.NewReader([]byte(content)))
	w.Close()

	return bytes.NewReader(buf.Bytes())
}

func testReadFirstLine(r io.Reader) (string, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}

	line, _, err := bufio.NewReader(gzipReader).ReadLine()
	if err != nil {
		return "", err
	}

	return string(line), nil
}
