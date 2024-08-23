// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Code generated by "pdata/internal/cmd/pdatagen/main.go". DO NOT EDIT.
// To regenerate this file run "make genpdata".

package ptrace

import (
	"sort"

	"go.opentelemetry.io/collector/pdata/internal"
	otlptrace "go.opentelemetry.io/collector/pdata/internal/data/protogen/trace/v1"
)

// SpanSlice logically represents a slice of Span.
//
// This is a reference type. If passed by value and callee modifies it, the
// caller will see the modification.
//
// Must use NewSpanSlice function to create new instances.
// Important: zero-initialized instance is not valid for use.
type SpanSlice struct {
	orig  *[]*otlptrace.Span
	state *internal.State
}

func newSpanSlice(orig *[]*otlptrace.Span, state *internal.State) SpanSlice {
	return SpanSlice{orig: orig, state: state}
}

// NewSpanSlice creates a SpanSlice with 0 elements.
// Can use "EnsureCapacity" to initialize with a given capacity.
func NewSpanSlice() SpanSlice {
	orig := []*otlptrace.Span(nil)
	state := internal.StateMutable
	return newSpanSlice(&orig, &state)
}

// Len returns the number of elements in the slice.
//
// Returns "0" for a newly instance created with "NewSpanSlice()".
func (es SpanSlice) Len() int {
	return len(*es.orig)
}

// At returns the element at the given index.
//
// This function is used mostly for iterating over all the values in the slice:
//
//	for i := 0; i < es.Len(); i++ {
//	    e := es.At(i)
//	    ... // Do something with the element
//	}
func (es SpanSlice) At(i int) Span {
	return newSpan((*es.orig)[i], es.state)
}

// EnsureCapacity is an operation that ensures the slice has at least the specified capacity.
// 1. If the newCap <= cap then no change in capacity.
// 2. If the newCap > cap then the slice capacity will be expanded to equal newCap.
//
// Here is how a new SpanSlice can be initialized:
//
//	es := NewSpanSlice()
//	es.EnsureCapacity(4)
//	for i := 0; i < 4; i++ {
//	    e := es.AppendEmpty()
//	    // Here should set all the values for e.
//	}
func (es SpanSlice) EnsureCapacity(newCap int) {
	es.state.AssertMutable()
	oldCap := cap(*es.orig)
	if newCap <= oldCap {
		return
	}

	newOrig := make([]*otlptrace.Span, len(*es.orig), newCap)
	copy(newOrig, *es.orig)
	*es.orig = newOrig
}

// AppendEmpty will append to the end of the slice an empty Span.
// It returns the newly added Span.
func (es SpanSlice) AppendEmpty() Span {
	es.state.AssertMutable()
	*es.orig = append(*es.orig, &otlptrace.Span{})
	return es.At(es.Len() - 1)
}

// MoveAndAppendTo moves all elements from the current slice and appends them to the dest.
// The current slice will be cleared.
func (es SpanSlice) MoveAndAppendTo(dest SpanSlice) {
	es.state.AssertMutable()
	dest.state.AssertMutable()
	if *dest.orig == nil {
		// We can simply move the entire vector and avoid any allocations.
		*dest.orig = *es.orig
	} else {
		*dest.orig = append(*dest.orig, *es.orig...)
	}
	*es.orig = nil
}

// RemoveIf calls f sequentially for each element present in the slice.
// If f returns true, the element is removed from the slice.
func (es SpanSlice) RemoveIf(f func(Span) bool) {
	es.state.AssertMutable()
	newLen := 0
	for i := 0; i < len(*es.orig); i++ {
		if f(es.At(i)) {
			continue
		}
		if newLen == i {
			// Nothing to move, element is at the right place.
			newLen++
			continue
		}
		(*es.orig)[newLen] = (*es.orig)[i]
		newLen++
	}
	*es.orig = (*es.orig)[:newLen]
}

// CopyTo copies all elements from the current slice overriding the destination.
func (es SpanSlice) CopyTo(dest SpanSlice) {
	dest.state.AssertMutable()
	srcLen := es.Len()
	destCap := cap(*dest.orig)
	if srcLen <= destCap {
		(*dest.orig) = (*dest.orig)[:srcLen:destCap]
		for i := range *es.orig {
			newSpan((*es.orig)[i], es.state).CopyTo(newSpan((*dest.orig)[i], dest.state))
		}
		return
	}
	origs := make([]otlptrace.Span, srcLen)
	wrappers := make([]*otlptrace.Span, srcLen)
	for i := range *es.orig {
		wrappers[i] = &origs[i]
		newSpan((*es.orig)[i], es.state).CopyTo(newSpan(wrappers[i], dest.state))
	}
	*dest.orig = wrappers
}

// Sort sorts the Span elements within SpanSlice given the
// provided less function so that two instances of SpanSlice
// can be compared.
func (es SpanSlice) Sort(less func(a, b Span) bool) {
	es.state.AssertMutable()
	sort.SliceStable(*es.orig, func(i, j int) bool { return less(es.At(i), es.At(j)) })
}
