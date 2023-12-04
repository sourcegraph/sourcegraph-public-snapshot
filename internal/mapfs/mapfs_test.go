package mapfs

import (
	"io"
	"io/fs"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMapFS(t *testing.T) {
	mapFS := New(map[string]string{
		"src/java/main/foo.java":                      "code",
		"src/scala/main/bar.scala":                    "better code",
		"test/java/main/baz.java":                     "qa",
		"test/scala/main/bonk.scala":                  "better qa",
		"business/slidedeck_2022_final~2_(1).pdf.bak": "stonks",
	})

	assertFile(t, mapFS, "src/scala/main/bar.scala", "better code")
	assertFile(t, mapFS, "test/scala/main/bonk.scala", "better qa")

	assertDirectory(t, mapFS, "", []string{"business", "src", "test"})
	assertDirectory(t, mapFS, "src", []string{"java", "scala"})
	assertDirectory(t, mapFS, "test", []string{"java", "scala"})
	assertDirectory(t, mapFS, "src/java", []string{"main"})
	assertDirectory(t, mapFS, "src/java/main", []string{"foo.java"})
}

func assertFile(t *testing.T, mapFS fs.FS, filename string, expectedContents string) {
	file, err := mapFS.Open(filename)
	if err != nil {
		t.Fatalf("failed to open file %q: %s", filename, err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	if diff := cmp.Diff(expectedContents, string(contents)); diff != "" {
		t.Fatalf("mismatched contents for file %q (-have, +want): %s", filename, diff)
	}
}

func assertDirectory(t *testing.T, mapFS fs.FS, directory string, expectedEntries []string) {
	file, err := mapFS.Open(directory)
	if err != nil {
		t.Fatalf("failed to open directory %q: %s", directory, err)
	}

	rdf, ok := file.(fs.ReadDirFile)
	if !ok {
		t.Fatalf("failed to read directory: bad type %T", file)
	}

	entries, err := rdf.ReadDir(-1)
	if err != nil {
		t.Fatalf("failed to read directory %q: %s", directory, err)
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	if diff := cmp.Diff(expectedEntries, names); diff != "" {
		t.Fatalf("mismatched entries for directory %q (-have, +want): %s", directory, diff)
	}
}
