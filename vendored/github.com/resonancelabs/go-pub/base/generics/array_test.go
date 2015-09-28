package generics

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestLen(t *testing.T) {

	testCases := []struct {
		input    interface{}
		expected int
	}{
		{[]int{}, 0},
		{[]int{0}, 1},
		{[]int{1}, 1},
		{[]int{1, 2, 3}, 3},
		{[]int{3, 3, 3}, 3},
		{[]string{}, 0},
		{[]string{"", "hello", " ", "world"}, 4},
		{[]string{"hello", "world"}, 2},
	}
	for _, c := range testCases {
		a := NewArrayFrom(c.input)
		if a.Len() != c.expected {
			t.Errorf("Incorrect length: found=%v, expected=%v", a.Len(), c.expected)
		}
	}
}

func TestGeneralOperations(t *testing.T) {

	a := NewArray()

	// Reduce boilerplate
	assertLen := func(expected int) {
		if a.Len() != expected {
			_, _, line, _ := runtime.Caller(1)
			t.Errorf("(line %d) Array length incorrect : found=%v, expected=%v", line, a.Len(), expected)
		}
	}
	assertValue := func(actual interface{}, expected interface{}) {
		if !reflect.DeepEqual(actual, expected) {
			_, _, line, _ := runtime.Caller(1)
			t.Errorf("(line %d) Value mismatch: found=%v, expected=%v", line, actual, expected)
		}
	}

	assertLen(0)

	a.Push(4)
	assertLen(1)
	fmt.Println(a)
	assertValue(a.Back(), 4)

	a.Push(42)
	assertLen(2)
	assertValue(a.Front(), 4)
	assertValue(a.Back(), 42)

	r0 := a.Pop().(int)
	assertValue(r0, 42)
	assertLen(1)
	assertValue(a.Front(), 4)
	assertValue(a.Back(), 4)

	a.Unshift(42)
	assertLen(2)
	assertValue(a.Front(), 42)
	assertValue(a.Back(), 4)

	r1 := a.Shift().(int)
	assertValue(r1, 42)
	assertLen(1)
	assertValue(a.Front(), 4)
	assertValue(a.Back(), 4)

	a.Unshift(3)
	a.Unshift(2)
	a.Unshift(1)
	a.Push(5)
	assertLen(5)
	assertValue(a.Front(), 1)
	assertValue(a.Back(), 5)

	a.Reverse()
	assertLen(5)
	assertValue(a.Front(), 5)
	assertValue(a.Back(), 1)

	a.First(3)
	assertLen(3)
	assertValue(a.Front(), 5)
	assertValue(a.Back(), 3)

	a.First(5)
	assertLen(3)
	assertValue(a.Front(), 5)
	assertValue(a.Back(), 3)

	a.Last(2)
	assertLen(2)
	assertValue(a.Front(), 4)
	assertValue(a.Back(), 3)

	a.Last(2000)
	assertLen(2)
	assertValue(a.Front(), 4)
	assertValue(a.Back(), 3)

	a.Last(0)
	assertLen(0)

	a.Push(1)
	a.Append(3, 5, 7, 9).Prepend(0, 2, 4, 6, 8)
	assertValue(a.Front(), 0)
	assertValue(a.Back(), 9)

	a.Sort(func(a, b int) bool {
		return b < a
	}).Reverse()
	assertValue(a.Front(), 0)
	assertValue(a.Back(), 9)
	assertValue(a.At(1), 1)
	assertValue(a.At(4), 4)

	a = NewArrayFrom([]int{0, 1, 2})
	a.AppendSlice([]int{3, 4, 5})
	assertValue(a.At(0), 0)
	assertValue(a.At(3), 3)
	assertValue(a.At(5), 5)

	a = NewArrayFrom([]int{0, 1, 2})
	a = a.Map(func(v int) int {
		return v * 2
	})
	assertValue(a.At(0), 0)
	assertValue(a.At(1), 2)
	assertValue(a.At(2), 4)

	a = NewArrayFrom([]int{0, 1, 2})
	a = a.Accept(func(v int) bool {
		return v%2 == 0
	})
	assertValue(a.Len(), 2)
	assertValue(a.At(0), 0)
	assertValue(a.At(1), 2)

	a = NewArrayFrom([]int{0, 1, 2})
	a = a.Reject(func(v int) bool {
		return v%2 == 0
	})
	assertValue(a.Len(), 1)
	assertValue(a.At(0), 1)

}

func TestSortPointerSlice(t *testing.T) {

	assertValue := func(actual interface{}, expected interface{}) {
		if !reflect.DeepEqual(actual, expected) {
			_, _, line, _ := runtime.Caller(1)
			t.Errorf("(line %d) Value mismatch: found=%v, expected=%v", line, actual, expected)
		}
	}

	type myType struct {
		Name  string
		Value int
	}

	slice := []*myType{
		&myType{"D", 4},
		&myType{"B", 2},
		&myType{"A", 1},
		&myType{"C", 3},
	}

	// Cast to slice
	sortedSlice := []*myType{}
	NewArrayFrom(slice).Sort(func(a, b *myType) bool {
		return a.Value < b.Value
	}).CopyToSlice(&sortedSlice)

	assertValue(sortedSlice[0].Name, "A")
	assertValue(sortedSlice[1].Name, "B")
	assertValue(sortedSlice[2].Name, "C")
	assertValue(sortedSlice[3].Name, "D")

	// Copy to new slice
	var sorted2 []*myType
	NewArrayFrom(slice).Sort(func(a, b *myType) bool {
		return a.Value < b.Value
	}).CopyToSlice(&sorted2)

	assertValue(sorted2[0].Name, "A")
	assertValue(sorted2[1].Name, "B")
	assertValue(sorted2[2].Name, "C")
	assertValue(sorted2[3].Name, "D")

	// Copy back to self
	NewArrayFrom(slice).Sort(func(a, b *myType) bool {
		return a.Value < b.Value
	}).CopyToSlice(&slice)

	assertValue(slice[0].Name, "A")
	assertValue(slice[1].Name, "B")
	assertValue(slice[2].Name, "C")
	assertValue(slice[3].Name, "D")

	var dstSlice []*myType

	var nilSlice []*myType = nil
	NewArrayFrom(nilSlice).Sort(func(a, b *myType) bool {
		return a.Value < b.Value
	}).CopyToSlice(&dstSlice)
	assertValue(len(dstSlice), 0)

	emptySlice := []*myType{}
	NewArrayFrom(emptySlice).Sort(func(a, b *myType) bool {
		return a.Value < b.Value
	}).CopyToSlice(&dstSlice)
	assertValue(len(dstSlice), 0)
}
