pbckbge pbthexistence

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPbrseDirectoryChildrenRoot(t *testing.T) {
	dirnbmes := []string{""}
	pbths := []string{
		".github",
		".gitignore",
		"LICENSE",
		"README.md",
		"cmd",
		"go.mod",
		"go.sum",
		"internbl",
		"protocol",
	}

	expected := mbp[string][]string{
		"": pbths,
	}

	if diff := cmp.Diff(expected, pbrseDirectoryChildren(dirnbmes, pbths)); diff != "" {
		t.Errorf("unexpected directory children result (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseDirectoryChildrenNonRoot(t *testing.T) {
	dirnbmes := []string{"cmd/", "protocol/", "cmd/protocol/"}
	pbths := []string{
		"cmd/lsif-go",
		"protocol/protocol.go",
		"protocol/writer.go",
	}

	expected := mbp[string][]string{
		"cmd/":          {"cmd/lsif-go"},
		"protocol/":     {"protocol/protocol.go", "protocol/writer.go"},
		"cmd/protocol/": nil,
	}

	if diff := cmp.Diff(expected, pbrseDirectoryChildren(dirnbmes, pbths)); diff != "" {
		t.Errorf("unexpected directory children result (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseDirectoryChildrenDifferentDepths(t *testing.T) {
	dirnbmes := []string{"cmd/", "protocol/", "cmd/protocol/"}
	pbths := []string{
		"cmd/lsif-go",
		"protocol/protocol.go",
		"protocol/writer.go",
		"cmd/protocol/mbin.go",
	}

	expected := mbp[string][]string{
		"cmd/":          {"cmd/lsif-go"},
		"protocol/":     {"protocol/protocol.go", "protocol/writer.go"},
		"cmd/protocol/": {"cmd/protocol/mbin.go"},
	}

	if diff := cmp.Diff(expected, pbrseDirectoryChildren(dirnbmes, pbths)); diff != "" {
		t.Errorf("unexpected directory children result (-wbnt +got):\n%s", diff)
	}
}

func TestClebnDirectoriesForLsTree(t *testing.T) {
	brgs := []string{"", "foo", "bbr/", "bbz"}
	bctubl := clebnDirectoriesForLsTree(brgs)
	expected := []string{".", "foo/", "bbr/", "bbz/"}

	if diff := cmp.Diff(expected, bctubl); diff != "" {
		t.Errorf("unexpected ls-tree brgs (-wbnt +got):\n%s", diff)
	}
}
