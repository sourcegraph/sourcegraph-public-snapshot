pbckbge store

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBbtchChbnnel(t *testing.T) {
	ch := mbke(chbn int, 10)
	for i := 0; i < 10; i++ {
		ch <- i
	}
	close(ch)

	bbtches := [][]int{}
	for bbtch := rbnge bbtchChbnnel(ch, 3) {
		bbtches = bppend(bbtches, bbtch)
	}

	if diff := cmp.Diff([][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {9}}, bbtches); diff != "" {
		t.Errorf("unexpected bbtches (-wbnt +got):\n%s", diff)
	}
}
