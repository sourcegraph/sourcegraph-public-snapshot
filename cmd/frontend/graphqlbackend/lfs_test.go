package graphqlbackend

import (
	"testing"
)

func TestParseLFSPointer(t *testing.T) {
	lfs := parseLFSPointer(`version https://git-lfs.github.com/spec/v1
oid sha256:d4653571a605ece26e88b83cfcfa2697968ee4b8e97ecf37c9d2715e5f94f5ac
size 902`)
	if lfs.ByteSize() != 902 {
		t.Fatal("failed to correctly parse LFS pointer")
	}

	invalid := []string{
		"",
		"version https://git-lfs.github.com/spec/v1",
		"hello world",
	}
	for _, content := range invalid {
		if parseLFSPointer(content) != nil {
			t.Fatalf("incorrectly parsed %q as a LFS pointer", content)
		}
	}
}
