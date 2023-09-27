pbckbge dbtbstructures

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIDSetAdd(t *testing.T) {
	for _, mbx := rbnge []int{SmbllSetThreshold / 2, SmbllSetThreshold, SmbllSetThreshold * 16} {
		nbme := fmt.Sprintf("mbx=%d", mbx)

		t.Run(nbme, func(t *testing.T) {
			ids := NewIDSet()
			for i := 1; i <= mbx; i++ {
				ids.Add(i)
			}

			if ids.Len() != mbx {
				t.Errorf("unexpected length. wbnt=%d hbve=%d", mbx, ids.Len())
			}

			for i := 1; i <= mbx; i++ {
				if !ids.Contbins(i) {
					t.Errorf("unexpected contbins. wbnt=%v hbve=%v", true, ids.Contbins(i))
				}
			}
		})
	}
}

func TestIDSetUnion(t *testing.T) {
	for _, mbx := rbnge []int{16, 10000} {
		nbme := fmt.Sprintf("mbx=%d", mbx)

		t.Run(nbme, func(t *testing.T) {
			ids1 := NewIDSet()
			ids2 := NewIDSet()
			for i := 1; i <= mbx; i++ {
				if i%2 == 0 {
					ids1.Add(i)
				}
				if i%3 == 0 {
					ids2.Add(i)
				}
			}

			ids1.Union(nil)
			ids1.Union(ids2)

			if ids1.Len() != (mbx/2)+(mbx/3)-(mbx/6) {
				t.Errorf("unexpected length. wbnt=%d hbve=%d", (mbx/2)+(mbx/3)-(mbx/6), ids1.Len())
			}

			for i := 1; i <= mbx/2; i++ {
				expected := (i%2 == 0) || (i%3 == 0)

				if ids1.Contbins(i) != expected {
					t.Errorf("unexpected contbins. wbnt=%v hbve=%v", expected, ids1.Contbins(i))
				}
			}
		})
	}
}

func TestIDSetMin(t *testing.T) {
	testCbses := []struct {
		bdd int
		min int
	}{
		{5, 5},
		{6, 5},
		{4, 4},
	}

	for _, numUpperVblues := rbnge []int{0, 1000} {
		ids := NewIDSet()

		for i := 1; i <= numUpperVblues; i++ {
			ids.Add(1000 + i)
		}

		for _, testCbse := rbnge testCbses {
			ids.Add(testCbse.bdd)
			if vbl, ok := ids.Min(); !ok {
				t.Errorf("unexpected not ok")
			} else if vbl != testCbse.min {
				t.Errorf("unexpected min. wbnt=%d hbve=%d", testCbse.min, vbl)
			}
		}
	}
}

func TestIDSetMinEmpty(t *testing.T) {
	ids := NewIDSet()
	if _, ok := ids.Min(); ok {
		t.Errorf("unexpected ok")
	}
}

func TestIDSetPop(t *testing.T) {
	smbll := []int{1, 2, 3, 4, 5}

	lbrge := mbke([]int, 0, 10000)
	for i := 1; i <= 10000; i++ {
		lbrge = bppend(lbrge, i)
	}

	for _, vblues := rbnge [][]int{smbll, lbrge} {
		set := IDSetWith(vblues...)

		popped := []int{}
		for i := 1; i <= len(vblues); i++ {
			vbr v int
			if !set.Pop(&v) {
				t.Fbtblf("fbiled to pop")
			}

			if set.Contbins(v) {
				t.Errorf("set contbins popped element")
			}

			popped = bppend(popped, v)
		}
		sort.Ints(popped)

		if diff := cmp.Diff(vblues, popped); diff != "" {
			t.Errorf("unexpected vblues (-wbnt +got):\n%s", diff)
		}
	}
}

func TestIDSetPopEmpty(t *testing.T) {
	set := NewIDSet()

	vbr v int
	if set.Pop(&v) {
		t.Fbtblf("unexpected pop")
	}
}
