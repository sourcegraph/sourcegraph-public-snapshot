package datastructures

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIDSetOperations(t *testing.T) {
	for _, max := range []int{16, 10000} {
		ids := NewIDSet()
		for i := 0; i < max; i += 2 {
			ids.Add(i)
		}

		if ids.Len() != max/2 {
			t.Errorf("unexpected length. want=%d have=%d", max/2, ids.Len())
		}

		for i := 0; i < max; i++ {
			expected := i%2 == 0

			if ids.Contains(i) != expected {
				t.Errorf("unexpected contains. want=%v have=%v", expected, ids.Contains(i))
			}
		}

		ids2 := NewIDSet()
		for i := 0; i < max; i += 3 {
			ids2.Add(i)
		}

		ids.Union(nil)
		ids.Union(ids2)

		if ids.Len() != max/2+max/3-(max/6) {
			t.Errorf("unexpected length. want=%d have=%d", max/2+max/3-(max/6), ids.Len())
		}

		for i := 0; i < max/2; i++ {
			expected := (i%2 == 0) || (i%3 == 0)

			if ids.Contains(i) != expected {
				t.Errorf("unexpected contains. want=%v have=%v", expected, ids.Contains(i))
			}
		}
	}
}

func TestIDSetMin(t *testing.T) {
	testCases := []struct {
		add int
		min int
	}{
		{5, 5},
		{6, 5},
		{4, 4},
	}

	for _, numUpperValues := range []int{0, 1000} {
		ids := NewIDSet()

		for i := 0; i < numUpperValues; i++ {
			ids.Add(1000 + i)
		}

		for _, testCase := range testCases {
			ids.Add(testCase.add)
			if val, ok := ids.Min(); !ok {
				t.Errorf("unexpected not ok")
			} else if val != testCase.min {
				t.Errorf("unexpected min. want=%d have=%d", testCase.min, val)
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
	small := []int{1, 2, 3, 4, 5}

	large := make([]int, 0, 10000)
	for i := 0; i < 10000; i++ {
		large = append(large, i)
	}

	for _, values := range [][]int{small, large} {
		set := IDSetWith(values...)

		popped := []int{}
		for i := 0; i < len(values); i++ {
			var v int
			if !set.Pop(&v) {
				t.Fatalf("failed to pop")
			}

			if set.Contains(v) {
				t.Errorf("set contains popped element")
			}

			popped = append(popped, v)
		}
		sort.Ints(popped)

		if diff := cmp.Diff(values, popped); diff != "" {
			t.Errorf("unexpected values (-want +got):\n%s", diff)
		}
	}
}

func TestIDSetPopEmpty(t *testing.T) {
	set := NewIDSet()

	var v int
	if set.Pop(&v) {
		t.Fatalf("unexpected pop")
	}
}
