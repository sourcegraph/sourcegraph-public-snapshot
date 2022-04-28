package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	"google.golang.org/protobuf/proto"
)

var exampleLsifTypedIndex = lsiftyped.Index{
	Metadata: &lsiftyped.Metadata{
		TextDocumentEncoding: lsiftyped.TextEncoding_UTF8,
		ToolInfo: &lsiftyped.ToolInfo{
			Name:    "hello",
			Version: "1.0.0",
		},
	},
}

var exampleLsifGraphString = `{"id":1,"version":"0.4.3","positionEncoding":"utf-8","toolInfo":{"name":"hello","version":"1.0.0"},"type":"vertex","label":"metaData"}
`

func exampleLsifTypedBytes(t *testing.T) []byte {
	bytes, err := proto.Marshal(&exampleLsifTypedIndex)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func createTempLsifTypedFile(t *testing.T) (typedFile, graphFile string) {
	dir := t.TempDir()
	typedFile = filepath.Join(dir, "dump.lsif-typed")
	graphFile = filepath.Join(dir, "dump.lsif")
	err := os.WriteFile(typedFile, exampleLsifTypedBytes(t), 0755)
	if err != nil {
		t.Fatal(err)
	}
	return typedFile, graphFile
}

func assertLsifGraphOutput(t *testing.T, lsifGraphFile, expectedGraphString string) {
	out := lsifUploadOutput()
	handleLSIFTyped(out)
	lsifGraph, err := os.ReadFile(lsifGraphFile)
	if err != nil {
		t.Fatal(err)
	}
	obtained := string(lsifGraph)
	if obtained != expectedGraphString {
		t.Fatalf("unexpected LSIF output %s", obtained)
	}
	if lsifGraphFile != lsifUploadFlags.file {
		t.Fatalf("unexpected lsifUploadFlag.file value %s, expected %s", lsifUploadFlags.file, lsifGraphFile)
	}
}

func TestImplicitlyConvertLsifTypedIntoGraph(t *testing.T) {
	_, graphFile := createTempLsifTypedFile(t)
	lsifUploadFlags.file = graphFile
	assertLsifGraphOutput(t, graphFile, exampleLsifGraphString)
}

func TestImplicitlyIgnoreLsifTyped(t *testing.T) {
	_, graphFile := createTempLsifTypedFile(t)
	lsifUploadFlags.file = graphFile
	os.WriteFile(graphFile, []byte("hello world"), 0755)
	assertLsifGraphOutput(t, graphFile, "hello world")
}

func TestExplicitlyConvertLsifTypedIntoGraph(t *testing.T) {
	typedFile, graphFile := createTempLsifTypedFile(t)
	lsifUploadFlags.file = typedFile
	assertLsifGraphOutput(t, graphFile, exampleLsifGraphString)
}
