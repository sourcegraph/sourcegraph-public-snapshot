package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

var exampleSCIPIndex = scip.Index{
	Metadata: &scip.Metadata{
		TextDocumentEncoding: scip.TextEncoding_UTF8,
		ToolInfo: &scip.ToolInfo{
			Name:    "hello",
			Version: "1.0.0",
		},
	},
}

var exampleLSIFString = `{"id":1,"version":"0.4.3","positionEncoding":"utf-8","toolInfo":{"name":"hello","version":"1.0.0"},"type":"vertex","label":"metaData"}
`

func exampleSCIPBytes(t *testing.T) []byte {
	bytes, err := proto.Marshal(&exampleSCIPIndex)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func createTempSCIPFile(t *testing.T, scipFileName string) (scipFilePath string, lsifFilePath string) {
	dir := t.TempDir()
	require.NotEqual(t, "", scipFileName)
	scipFilePath = filepath.Join(dir, scipFileName)
	lsifFilePath = filepath.Join(dir, "dump.lsif")
	err := os.WriteFile(scipFilePath, exampleSCIPBytes(t), 0755)
	if err != nil {
		t.Fatal(err)
	}
	return scipFilePath, lsifFilePath
}

func assertLSIFOutput(t *testing.T, lsifFile, expectedLSIFString string) {
	out := codeintelUploadOutput()
	handleSCIP(out)
	lsif, err := os.ReadFile(lsifFile)
	if err != nil {
		t.Fatal(err)
	}
	obtained := string(lsif)
	if obtained != expectedLSIFString {
		t.Fatalf("unexpected LSIF output %s", obtained)
	}
	if lsifFile != codeintelUploadFlags.file {
		t.Fatalf("unexpected codeintelUploadFlag.file value %s, expected %s", codeintelUploadFlags.file, lsifFile)
	}
}

func TestImplicitlyConvertSCIPIntoLSIF(t *testing.T) {
	for _, filename := range []string{"index.scip", "dump.scip", "dump.lsif-typed"} {
		_, lsifFile := createTempSCIPFile(t, filename)
		codeintelUploadFlags.file = lsifFile
		assertLSIFOutput(t, lsifFile, exampleLSIFString)
	}
}

func TestImplicitlyIgnoreSCIP(t *testing.T) {
	for _, filename := range []string{"index.scip", "dump.scip", "dump.lsif-typed"} {
		_, lsifFile := createTempSCIPFile(t, filename)
		codeintelUploadFlags.file = lsifFile
		os.WriteFile(lsifFile, []byte("hello world"), 0755)
		assertLSIFOutput(t, lsifFile, "hello world")
	}
}

func TestExplicitlyConvertSCIPIntoGraph(t *testing.T) {
	for _, filename := range []string{"index.scip", "dump.scip", "dump.lsif-typed"} {
		scipFile, lsifFile := createTempSCIPFile(t, filename)
		codeintelUploadFlags.file = scipFile
		assertLSIFOutput(t, lsifFile, exampleLSIFString)
	}
}

func TestReplaceExtension(t *testing.T) {
	require.Panics(t, func() { replaceExtension("foo", ".xyz") })
	require.Equal(t, "foo.xyz", replaceExtension("foo.abc", ".xyz"))
}

func TestReplaceBaseName(t *testing.T) {
	require.Panics(t, func() { replaceBaseName("mydir", filepath.Join("dir", "file")) })

	require.Equal(t, filepath.Join("a", "d.e"),
		replaceBaseName(filepath.Join("a", "b.c"), "d.e"))
}
