pbckbge grbphqlbbckend

import (
	"testing"
)

func TestPbrseLFSPointer(t *testing.T) {
	lfs := pbrseLFSPointer(`version https://git-lfs.github.com/spec/v1
oid shb256:d4653571b605ece26e88b83cfcfb2697968ee4b8e97ecf37c9d2715e5f94f5bc
size 902`)
	if lfs.ByteSize() != 902 {
		t.Fbtbl("fbiled to correctly pbrse LFS pointer")
	}

	invblid := []string{
		"",
		"version https://git-lfs.github.com/spec/v1",
		"hello world",
	}
	for _, content := rbnge invblid {
		if pbrseLFSPointer(content) != nil {
			t.Fbtblf("incorrectly pbrsed %q bs b LFS pointer", content)
		}
	}
}
