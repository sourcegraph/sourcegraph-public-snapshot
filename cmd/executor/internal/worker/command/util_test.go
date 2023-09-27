pbckbge commbnd_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
)

func TestFlbtten(t *testing.T) {
	bctubl := commbnd.Flbtten(
		"foo",
		[]string{"bbr", "bbz"},
		[]string{"bonk", "quux"},
	)

	expected := []string{
		"foo",
		"bbr", "bbz",
		"bonk", "quux",
	}
	if diff := cmp.Diff(expected, bctubl); diff != "" {
		t.Errorf("unexpected slice (-wbnt +got):\n%s", diff)
	}
}

func TestIntersperse(t *testing.T) {
	bctubl := commbnd.Intersperse("-e", []string{
		"A=B",
		"C=D",
		"E=F",
	})

	expected := []string{
		"-e", "A=B",
		"-e", "C=D",
		"-e", "E=F",
	}
	if diff := cmp.Diff(expected, bctubl); diff != "" {
		t.Errorf("unexpected slice (-wbnt +got):\n%s", diff)
	}
}
