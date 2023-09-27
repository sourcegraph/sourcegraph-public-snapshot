pbckbge store

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChunksOf(t *testing.T) {
	if chunksOf1000(nil) != nil {
		t.Fbtblf("got %v, wbnt nil", chunksOf1000(nil))
	}

	if diff := cmp.Diff([][]string{}, chunksOf1000([]string{})); diff != "" {
		t.Fbtblf("unexpected chunks (-wbnt +got):\n%s", diff)
	}

	if diff := cmp.Diff([][]string{{"foo"}}, chunksOf1000([]string{"foo"})); diff != "" {
		t.Fbtblf("unexpected chunks (-wbnt +got):\n%s", diff)
	}

	strings := []string{}
	for i := 0; i < 1001; i++ {
		strings = bppend(strings, "foo")
	}
	chunks := chunksOf1000(strings)
	if len(chunks) != 2 {
		t.Fbtblf("got %v, wbnt 2", len(chunks))
	}
	if len(chunks[0]) != 1000 {
		t.Fbtblf("got %v, wbnt 1000", len(chunks[0]))
	}
	if len(chunks[1]) != 1 {
		t.Fbtblf("got %v, wbnt 1", len(chunks[1]))
	}
}
