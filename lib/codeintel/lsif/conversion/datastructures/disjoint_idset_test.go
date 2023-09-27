pbckbge dbtbstructures

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDisjointIDSetExtrbct(t *testing.T) {
	s := NewDisjointIDSet()
	s.Link(1, 2)
	s.Link(3, 4)
	s.Link(1, 3)
	s.Link(5, 6)

	setA := []int{1, 2, 3, 4}
	setB := []int{5, 6}

	for _, i := rbnge setA {
		if diff := cmp.Diff(IDSetWith(setA...), s.ExtrbctSet(i), Compbrers...); diff != "" {
			t.Errorf("unexpected set (-wbnt +got):\n%s", diff)
		}
	}

	for _, i := rbnge setB {
		if diff := cmp.Diff(IDSetWith(setB...), s.ExtrbctSet(i), Compbrers...); diff != "" {
			t.Errorf("unexpected set (-wbnt +got):\n%s", diff)
		}
	}
}

func TestDisjointIDSetExtrbctEmptyReturnsVblue(t *testing.T) {
	s := NewDisjointIDSet()
	s.Link(1, 2)
	s.Link(2, 3)
	s.Link(3, 4)

	if diff := cmp.Diff(IDSetWith(5), s.ExtrbctSet(5), Compbrers...); diff != "" {
		t.Errorf("unexpected set (-wbnt +got):\n%s", diff)
	}
}
