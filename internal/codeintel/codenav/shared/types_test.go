package shared

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestUnitComparePosition(t *testing.T) {
	positions := []Position{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
		{1, 2},
		{2, 1},
		{2, 2},
	}

	type testCase struct {
		p1       Position
		p2       Position
		expected int
	}

	testCases := make([]testCase, len(positions)*len(positions))

	// There is a total ordering on these positions, so we just test all possible combinations
	for i1, pos1 := range positions {
		for i2, pos2 := range positions {
			testCases = append(testCases, testCase{pos1, pos2, cmp.Compare(i1, i2)})
		}
	}

	for _, testCase := range testCases {
		if actual := testCase.p1.Compare(testCase.p2); actual != testCase.expected {
			t.Errorf("unexpected result. %+v.Compare(%+v) want=%d have=%d", testCase.p1, testCase.p2, testCase.expected, actual)
		}
	}
}

func TestUnitIntersect(t *testing.T) {
	r1 := Range{Position{0, 0}, Position{0, 10}}
	r2 := Range{Position{0, 5}, Position{1, 5}}
	r3 := Range{Position{1, 0}, Position{1, 10}}

	require.Truef(t, r1.Intersects(r2), "%+v.Intersects(%+v)", r1, r2)
	require.Truef(t, r2.Intersects(r3), "%+v.Intersects(%+v)", r2, r3)
	require.Falsef(t, r1.Intersects(r3), "%+v.Intersects(%+v)", r1, r3)
}

func genPosition() *rapid.Generator[Position] {
	return rapid.Custom(func(t *rapid.T) Position {
		return Position{
			Line:      rapid.IntRange(0, 10).Draw(t, "Line"),
			Character: rapid.IntRange(0, 10).Draw(t, "Character"),
		}
	})
}

func genRange() *rapid.Generator[Range] {
	return rapid.Custom(func(t *rapid.T) Range {
		posGen := genPosition()
		start := posGen.Draw(t, "start")
		end := posGen.Draw(t, "end")
		// A well formed range is always start <= end.
		if start.Compare(end) > 0 {
			start, end = end, start
		}
		return Range{
			Start: start,
			End:   end,
		}
	})
}

func TestComparePositionTransitive(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p1 := genPosition().Draw(t, "p1")
		p2 := genPosition().Draw(t, "p2")
		p3 := genPosition().Draw(t, "p3")

		if p1.Compare(p2) < 0 && p2.Compare(p3) < 0 {
			if !(p1.Compare(p3) < 0) {
				t.Errorf("%+v < %+v < %+v but !(%+v < %+v)", p1, p2, p3, p1, p3)
			}
		}
	})
}

func TestIntersects(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		r1 := genRange().Draw(t, "r1")
		r2 := genRange().Draw(t, "r2")

		if r1.Intersects(r2) {
			if !(r1.Contains(r2.Start) || r2.Contains(r1.Start)) {
				t.Errorf("%+v overlaps with %+v but neither contains the others start position", r1, r2)
			}
		}
	})
}
