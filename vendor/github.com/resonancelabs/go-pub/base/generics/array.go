package generics

import (
	"reflect"
	"sort"
)

// A generic array to avoid the boilerplate of strongly typed Go slices.
// Intended for cases where convenience of coding is preferable to strict
// typing and runtime efficiency -- e.g. proof of concept code.
//
// Ideally, Array would provide functional equivalents of all the conveniences
// of a library like underscore.js (with adjustments for the different language
// semantics).
//
// Also see https://github.com/bradfitz/slice for likely faster generic sorting.
type Array struct {
	data []interface{}
}

func NewArray() *Array {
	return &Array{make([]interface{}, 0)}
}

// iface should be a slice of any any element type.  Will create a generic
// Array object wrapper around that data.  Will panic() in the case a non-slice
// is passed into the function.
func NewArrayFrom(iface interface{}) *Array {

	arr := reflect.ValueOf(iface)
	if arr.Kind() != reflect.Slice {
		panic("NewArrayFrom called with non-slice interface{}")
	}
	length := arr.Len()

	p := &Array{}
	p.data = make([]interface{}, length)
	for i := 0; i < length; i++ {
		p.data[i] = arr.Index(i).Interface()
	}
	return p
}

func (p Array) Len() int {
	return len(p.data)
}

// Adds an element to the end of the slice.
func (p *Array) Push(e interface{}) {
	p.data = append(p.data, e)
}

// Removes an element from the end of the slice and returns it.
func (p *Array) Pop() interface{} {
	e := p.data[len(p.data)-1]
	p.data = p.data[:len(p.data)-1]
	return e
}

// Adds an element to the front of the slice.
func (p *Array) Unshift(e interface{}) {
	p.data = append([]interface{}{e}, p.data...)
}

// Removes and returns an element from the front of the slice.
func (p *Array) Shift() interface{} {
	if len(p.data) == 0 {
		return nil
	}
	e := p.data[0]
	p.data = p.data[1:]
	return e
}

func (p Array) At(i int) interface{} {
	return p.data[i]
}

func (p Array) Front() interface{} {
	if len(p.data) == 0 {
		return nil
	} else {
		return p.data[0]
	}
}

func (p Array) Back() interface{} {
	if len(p.data) == 0 {
		return nil
	} else {
		return p.data[len(p.data)-1]
	}
}

func (p *Array) Append(elements ...interface{}) *Array {
	p.data = append(p.data, elements...)
	return p
}

func (p *Array) Prepend(elements ...interface{}) *Array {
	p.data = append(elements, p.data...)
	return p
}

func (p *Array) AppendSlice(s interface{}) {
	q := NewArrayFrom(s)
	p.Append(q.data...)
}

func (p *Array) PrependSlice(s interface{}) {
	q := NewArrayFrom(s)
	p.Prepend(q.data...)
}

// Reduce the internal array size to the first N elements. If there are
// already N or less items, the array size is unchanged.
func (p *Array) First(limit int) *Array {
	if len(p.data) > limit {
		p.data = p.data[:limit]
	}
	return p
}

// Reduce the internal array length to the last N elements. If there are
// already N or less items, the array size is unchanged.
func (p *Array) Last(limit int) *Array {
	if len(p.data) > limit {
		p.data = p.data[len(p.data)-limit:]
	}
	return p
}

// Reverse the order of the internal array
func (p *Array) Reverse() *Array {
	for i := 0; i < len(p.data)/2; i++ {
		j := len(p.data) - i - 1
		p.data[i], p.data[j] = p.data[j], p.data[i]
	}
	return p
}

type interfaceSlice []interface{}

func (s interfaceSlice) Len() int      { return len(s) }
func (s interfaceSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortableInterfaceSlice struct {
	interfaceSlice
	lessFunc func(a, b interface{}) bool
}

func (s sortableInterfaceSlice) Less(i, j int) bool {
	return s.lessFunc(s.interfaceSlice[i], s.interfaceSlice[j])
}

// Sorts the internal contents of the Array via a function.  Typical
// implementations will cast the data to the expected type and do the
// comparison. E.g.:
//
//	generics.NewArrayFrom([]int{5, 2, 8, 19, 3}).SortCallback(func(a, b interface{}) bool {
//		return a.(int) < b.(int)
//	}).CopyToSlice(&myIntSlice)
//	fmt.Println(sorted)
//
func (p *Array) SortCallback(lessFunc func(a, b interface{}) bool) *Array {
	sorted := sortableInterfaceSlice{
		interfaceSlice: p.data,
		lessFunc:       lessFunc,
	}
	sort.Sort(sorted)
	return p
}

// At the expense of compile-time type checking and some runtime performance,
// provides a generic Sort without casting:
//
//	sorted := generics.NewArrayFrom([]int{5, 2, 8, 19, 3}).Sort(func(a, b int) bool {
//		return a < b
//	}).CopyToSlice(&myIntSlice)
//	fmt.Println(sorted)
//
func (p *Array) Sort(lessFunc interface{}) *Array {
	cb := reflect.ValueOf(lessFunc)
	wrapper := func(a, b interface{}) bool {
		r := cb.Call([]reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)})
		return r[0].Interface().(bool)
	}
	return p.SortCallback(wrapper)
}

// Make a new array via an element-by-element transformation.
func (p *Array) mapWithCallback(callback func(interface{}) interface{}) *Array {
	q := &Array{data: make([]interface{}, len(p.data))}
	for i, e := range p.data {
		q.data[i] = callback(e)
	}
	return q
}

// Return a new array of transformed elements. The callback should be of the
// form:
//
// func(v valueType) valueType
//
func (p *Array) Map(mapFunc interface{}) *Array {
	cb := reflect.ValueOf(mapFunc)
	wrapper := func(v interface{}) interface{} {
		r := cb.Call([]reflect.Value{reflect.ValueOf(v)})
		return r[0].Interface()
	}
	return p.mapWithCallback(wrapper)
}

func (p *Array) acceptWithCallback(callback func(interface{}) bool) *Array {
	q := NewArray()
	for _, e := range p.data {
		if callback(e) {
			q.Push(e)
		}
	}
	return q
}

// Returns a new array with only the elements passing the acceptance callback.
//
// func (v valueType) bool
//
func (p *Array) Accept(acceptFunc interface{}) *Array {
	cb := reflect.ValueOf(acceptFunc)
	wrapper := func(v interface{}) bool {
		r := cb.Call([]reflect.Value{reflect.ValueOf(v)})
		return r[0].Interface().(bool)
	}
	return p.acceptWithCallback(wrapper)
}

// Returns a new array with the elements that return true from rejection test
// removed.
//
// func (v valueType) bool
//
func (p *Array) Reject(rejectFunc interface{}) *Array {
	cb := reflect.ValueOf(rejectFunc)
	wrapper := func(v interface{}) bool {
		r := cb.Call([]reflect.Value{reflect.ValueOf(v)})
		return !r[0].Interface().(bool)
	}
	return p.acceptWithCallback(wrapper)
}

// Return the internal slice of interface{} elements.
func (p Array) Data() []interface{} {
	return p.data
}

// REMOVED: ToSlice() interface{}
//
//
// Removed as the empty array case has to be special-cased  by the caller as
// it returns a nil interface{}, not an empty slice of the correct type. This
// lead to possibilities of panic()'s due to bad casts when the empty array
// case was not considered and is not (trivially) fixable in the method itself
// as there is not sufficient type information in the empty case to create a
// correctly typed slice.
//
// ---
//
// Return the internal slice as an interface{} that can be safely and correctly
// cast to a typed slice.
//
// Presumes that the type of the elements is uniform across the array.  This is
// not necessarily a requirement of other methods of Array.
//
// Returns nil in the case of an empty internal array.
//
/*func (p Array) ToSlice() interface{} {

	// Need to return a nil slice in the case of no data since there's no data
	// to infer the destination type from.
	if len(p.data) == 0 {
		return nil
	}

	elemType := reflect.TypeOf(p.data[0])
	sliceType := reflect.SliceOf(elemType)
	slice := reflect.MakeSlice(sliceType, len(p.data), len(p.data))
	for i := 0; i < len(p.data); i++ {
		slice.Index(i).Set(reflect.ValueOf(p.data[i]))
	}
	return slice.Interface()
}*/

// Copies the internal slice to the given destination slice.  The destination
// should be a *pointer* to a slice of the correct type.
//
// Unlike ToSlice(), this will set the result to a correctly typed empty slice
// rather than nil if the internal array is empty.
//
// Example:
//
// type myType struct {
// 	  Name  string
// 	  Value int
// }
//
// slice := []*myType{
// 	  &myType{"D", 4},
// 	  &myType{"B", 2},
// 	  &myType{"A", 1},
// 	  &myType{"C", 3},
// }
//
// NewArrayFrom(slice).Sort(func(a, b *myType) bool {
//   	return a.Value < b.Value
// }).CopyToSlice(&slice)
//
func (p Array) CopyToSlice(dst interface{}) {
	arrPtr := reflect.ValueOf(dst)
	if arrPtr.Kind() != reflect.Ptr {
		panic("CopyToSlice must be called with a pointer to a slice")
	}
	arr := arrPtr.Elem()
	if arr.Kind() != reflect.Slice {
		panic("CopyToSlice must be called with a pointer to a slice")
	}

	slice := reflect.MakeSlice(arr.Type(), len(p.data), len(p.data))
	for i := 0; i < len(p.data); i++ {
		slice.Index(i).Set(reflect.ValueOf(p.data[i]))
	}
	arrPtr.Elem().Set(slice)
}
