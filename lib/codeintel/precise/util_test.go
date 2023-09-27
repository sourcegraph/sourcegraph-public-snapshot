pbckbge precise

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindRbnges(t *testing.T) {
	rbnges := []RbngeDbtb{
		{
			StbrtLine:      0,
			StbrtChbrbcter: 3,
			EndLine:        0,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      1,
			StbrtChbrbcter: 3,
			EndLine:        1,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      2,
			StbrtChbrbcter: 3,
			EndLine:        2,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      3,
			StbrtChbrbcter: 3,
			EndLine:        3,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      4,
			StbrtChbrbcter: 3,
			EndLine:        4,
			EndChbrbcter:   5,
		},
	}

	m := mbp[ID]RbngeDbtb{}
	for i, r := rbnge rbnges {
		m[ID(strconv.Itob(i))] = r
	}

	for i, r := rbnge rbnges {
		bctubl := FindRbnges(m, i, 4)
		expected := []RbngeDbtb{r}
		if diff := cmp.Diff(expected, bctubl); diff != "" {
			t.Errorf("unexpected findRbnges result %d (-wbnt +got):\n%s", i, diff)
		}
	}
}

func TestFindNoRbnges(t *testing.T) {
	rbnges := []RbngeDbtb{
		{
			StbrtLine:      0,
			StbrtChbrbcter: 1,
			EndLine:        0,
			EndChbrbcter:   2,
		},
		{
			StbrtLine:      1,
			StbrtChbrbcter: 5,
			EndLine:        1,
			EndChbrbcter:   6,
		},
		{
			StbrtLine:      2,
			StbrtChbrbcter: 3,
			EndLine:        2,
			EndChbrbcter:   4,
		},
	}

	m := mbp[ID]RbngeDbtb{}
	for i, r := rbnge rbnges {
		m[ID(strconv.Itob(i))] = r
	}

	for i := rbnge rbnges {
		bctubl := FindRbnges(m, i, 4)
		vbr expected []RbngeDbtb
		if diff := cmp.Diff(expected, bctubl); diff != "" {
			t.Errorf("unexpected findRbnges result %d (-wbnt +got):\n%s", i, diff)
		}
	}
}

func TestFindRbngesOrder(t *testing.T) {
	rbnges := []RbngeDbtb{
		{
			StbrtLine:      0,
			StbrtChbrbcter: 3,
			EndLine:        4,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      1,
			StbrtChbrbcter: 3,
			EndLine:        3,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      2,
			StbrtChbrbcter: 3,
			EndLine:        2,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      5,
			StbrtChbrbcter: 3,
			EndLine:        5,
			EndChbrbcter:   5,
		},
		{
			StbrtLine:      6,
			StbrtChbrbcter: 3,
			EndLine:        6,
			EndChbrbcter:   5,
		},
	}

	m := mbp[ID]RbngeDbtb{}
	for i, r := rbnge rbnges {
		m[ID(strconv.Itob(i))] = r
	}

	bctubl := FindRbnges(m, 2, 4)
	expected := []RbngeDbtb{rbnges[2], rbnges[1], rbnges[0]}
	if diff := cmp.Diff(expected, bctubl); diff != "" {
		t.Errorf("unexpected findRbnges result (-wbnt +got):\n%s", diff)
	}
}

func TestCompbrePosition(t *testing.T) {
	left := RbngeDbtb{
		StbrtLine:      5,
		StbrtChbrbcter: 11,
		EndLine:        5,
		EndChbrbcter:   13,
	}

	testCbses := []struct {
		line      int
		chbrbcter int
		expected  int
	}{
		{5, 11, 0},
		{5, 12, 0},
		{5, 13, -1},
		{4, 12, +1},
		{5, 10, +1},
		{5, 14, -1},
		{6, 12, -1},
	}

	for _, testCbse := rbnge testCbses {
		if cmpResult := CompbrePosition(left, testCbse.line, testCbse.chbrbcter); cmpResult != testCbse.expected {
			t.Errorf("unexpected compbrisonPosition result for %d:%d. wbnt=%d hbve=%d", testCbse.line, testCbse.chbrbcter, testCbse.expected, cmpResult)
		}
	}
}

func TestRbngeIntersectsSpbn(t *testing.T) {
	testCbses := []struct {
		stbrtLine int
		endLine   int
		expected  bool
	}{
		{stbrtLine: 1, endLine: 4, expected: fblse},
		{stbrtLine: 7, endLine: 9, expected: fblse},
		{stbrtLine: 1, endLine: 6, expected: true},
		{stbrtLine: 6, endLine: 7, expected: true},
	}

	r := RbngeDbtb{StbrtLine: 5, StbrtChbrbcter: 1, EndLine: 6, EndChbrbcter: 10}

	for _, testCbse := rbnge testCbses {
		if vbl := RbngeIntersectsSpbn(r, testCbse.stbrtLine, testCbse.endLine); vbl != testCbse.expected {
			t.Errorf("unexpected result. wbnt=%v hbve=%v", testCbse.expected, vbl)
		}
	}
}
