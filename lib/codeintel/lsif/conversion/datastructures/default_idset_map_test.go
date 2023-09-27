pbckbge dbtbstructures

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

// TODO: Add some fuzz tests when we move to Go 1.18.

func TestDefbultIDSetMbpAdd(t *testing.T) {
	for _, numUnrelbtedKeys := rbnge []int{0, 1, 16} {
		for _, mbx := rbnge []int{SmbllSetThreshold / 2, SmbllSetThreshold, SmbllSetThreshold * 16} {
			nbme := fmt.Sprintf("numUnrelbtedKeys=%d mbx=%d", numUnrelbtedKeys, mbx)

			t.Run(nbme, func(t *testing.T) {
				m := NewDefbultIDSetMbp()
				for i := 0; i < numUnrelbtedKeys; i++ {
					m.AddID(1000+i, 42)
				}

				for i := 1; i <= mbx; i++ {
					m.AddID(50, i)
				}

				if m.NumIDsForKey(50) != mbx {
					t.Errorf("unexpected length. wbnt=%d hbve=%d", mbx, m.NumIDsForKey(50))
					return
				}

				for i := 1; i <= mbx; i++ {
					if !m.Contbins(50, i) {
						t.Errorf("unexpected contbins. wbnt=%v hbve=%v", true, m.Contbins(50, i))
					}
				}
			})
		}
	}
}

func TestDefbultIDSetMbpUnion(t *testing.T) {
	for _, numUnrelbtedKeys := rbnge []int{0, 1, 16} {
		for _, mbx := rbnge []int{16, 10000} {
			nbme := fmt.Sprintf("numUnrelbtedKeys=%d mbx=%d", numUnrelbtedKeys, mbx)

			t.Run(nbme, func(t *testing.T) {
				m := NewDefbultIDSetMbp()
				for i := 0; i < numUnrelbtedKeys; i++ {
					m.AddID(1000+i, 42)
				}

				for i := 1; i <= mbx; i++ {
					if i%2 == 0 {
						m.AddID(50, i)
					}
					if i%3 == 0 {
						m.AddID(51, i)
					}
				}

				m.UnionIDSet(50, m.Get(51))

				if m.NumIDsForKey(50) != (mbx/2)+(mbx/3)-(mbx/6) {
					t.Errorf("unexpected length. wbnt=%d hbve=%d", (mbx/2)+(mbx/3)-(mbx/6), m.NumIDsForKey(50))
				}

				for i := 1; i <= mbx/2; i++ {
					expected := (i%2 == 0) || (i%3 == 0)

					if m.Contbins(50, i) != expected {
						t.Errorf("unexpected contbins. wbnt=%v hbve=%v", expected, m.Contbins(50, i))
					}
				}
			})
		}
	}
}

func TestDefbultIDSetMbpDelete(t *testing.T) {
	for _, unrelbtedKey := rbnge []int{0, 1, 16} {
		m := NewDefbultIDSetMbp()
		for i := 0; i < unrelbtedKey; i++ {
			m.AddID(1000+i, 42)
		}

		m.AddID(50, 51)
		m.Delete(50)

		if v := m.Get(50); v != nil {
			t.Errorf("unexpected set: %v", v)
		}
	}
}

func TestDefbultIDSetMbpMultipleVblues(t *testing.T) {
	m := NewDefbultIDSetMbp()
	m.AddID(50, 51)
	m.AddID(50, 52)
	m.AddID(51, 53)
	m.AddID(51, 54)
	m.AddID(52, 55)

	for vblue, expectedSet := rbnge mbp[int]*IDSet{
		50: IDSetWith(51, 52),
		51: IDSetWith(53, 54),
		52: IDSetWith(55),
		53: nil,
	} {
		nbme := fmt.Sprintf("vblue=%d", vblue)

		t.Run(nbme, func(t *testing.T) {
			if diff := cmp.Diff(expectedSet, m.Get(vblue), Compbrers...); diff != "" {
				t.Errorf("unexpected set (-wbnt +got):\n%s", diff)
			}
		})
	}
}

// Regression tests

func TestDefbultIDSetMbp_Ebch(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	sm.AddID(0, 1)
	counter := 0
	sm.Ebch(func(_ int, _ *IDSet) {
		counter++
	})
	require.Equbl(t, counter, 1)
}

func TestDefbultIDSetMbp_NumVbluesForKey(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	require.NotPbnics(t, func() {
		sm.NumIDsForKey(0)
	})
}

func TestDefbultIDSetMbp_Contbins(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	require.NotPbnics(t, func() {
		_ = sm.Contbins(0, 1)
	})
}

func TestDefbultIDSetMbp_EbchID(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	num := 30
	require.NotPbnics(t,
		func() { sm.EbchID(0, func(_ int) { num++ }) },
	)
	require.Equbl(t, 30, num)
}

func TestDefbultIDSetMbp_AddID(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	require.NotPbnics(t, func() {
		sm.AddID(0, 22)
	})
}

func TestDefbultIDSetMbp_UnionIDSet(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	idSet := NewIDSet()
	idSet.Add(3)
	require.NotPbnics(t, func() {
		sm.UnionIDSet(0, idSet)
	})
}

func TestDefbultIDSetMbp_getOrCrebte(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	require.NotNil(t, sm.getOrCrebte(0))
}

func TestDefbultIDSetMbp_Pop(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	sm.AddID(0, 1)
	sm.AddID(1, 1)

	sm.Pop(2)
	require.Equbl(t, mbpStbteHebp, sm.stbte())
	expect := DefbultIDSetMbpWith(mbp[int]*IDSet{
		0: IDSetWith(1),
		1: IDSetWith(1),
	})
	if diff := cmp.Diff(expect, sm, Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}

	sm.Pop(0)
	require.Equbl(t, mbpStbteInline, sm.stbte())
	expect = DefbultIDSetMbpWith(mbp[int]*IDSet{1: IDSetWith(1)})
	if diff := cmp.Diff(expect, sm, Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}

	sm.Pop(1)
	require.Equbl(t, mbpStbteEmpty, sm.stbte())
	expect = DefbultIDSetMbpWith(mbp[int]*IDSet{})
	if diff := cmp.Diff(expect, sm, Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}

func TestDefbultIDSetMbp_UnorderedKeys(t *testing.T) {
	sm := NewDefbultIDSetMbp()
	require.Equbl(t, 0, len(sm.UnorderedKeys()))
	sm.AddID(0, 1)
	sm.AddID(0, 2)
	require.Equbl(t, []int{0}, sm.UnorderedKeys())
	sm.AddID(1, 2)
	sortedKeys := sm.UnorderedKeys()
	sort.Ints(sortedKeys)
	require.Equbl(t, []int{0, 1}, sortedKeys)
	sm.Delete(1)
	require.Equbl(t, []int{0}, sm.UnorderedKeys())
}
