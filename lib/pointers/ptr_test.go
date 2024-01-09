package pointers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPtr(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
	}{
		{
			name: "int",
			val:  1,
		},
		{
			name: "string",
			val:  "hello",
		},
		{
			name: "bool",
			val:  true,
		},
		{
			name: "struct",
			val:  struct{ Foo int }{42},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ptr(tt.val)
			assert.Equal(t, tt.val, *got)
		})
	}
}

type nonZeroTestCase[T comparable] struct {
	name    string
	val     T
	wantNil bool
}

func runNonZeroPtrTest[T comparable](t *testing.T, tc nonZeroTestCase[T]) {
	t.Helper()

	t.Run(tc.name, func(t *testing.T) {
		got := NonZeroPtr(tc.val)
		if tc.wantNil {
			require.Nil(t, got)
		} else {
			assert.Equal(t, tc.val, *got)
		}
	})
}

func TestNonZeroPtr(t *testing.T) {
	intTests := []nonZeroTestCase[int]{
		{
			name: "int",
			val:  1,
		},
		{
			name:    "zero int",
			val:     0,
			wantNil: true,
		},
	}
	stringTests := []nonZeroTestCase[string]{
		{
			name: "string",
			val:  "hello",
		},
		{
			name:    "zero string",
			val:     "",
			wantNil: true,
		},
	}
	boolTests := []nonZeroTestCase[bool]{
		{
			name: "bool",
			val:  true,
		},
		{
			name:    "zero bool",
			val:     false,
			wantNil: true,
		},
	}
	structTests := []nonZeroTestCase[struct{ Foo int }]{
		{
			name: "struct",
			val:  struct{ Foo int }{42},
		},
		{
			name:    "zero struct",
			val:     struct{ Foo int }{},
			wantNil: true,
		},
	}

	for _, tc := range intTests {
		runNonZeroPtrTest(t, tc)
	}
	for _, tc := range stringTests {
		runNonZeroPtrTest(t, tc)
	}
	for _, tc := range boolTests {
		runNonZeroPtrTest(t, tc)
	}
	for _, tc := range structTests {
		runNonZeroPtrTest(t, tc)
	}
}

type derefTestCase[T comparable] struct {
	name       string
	val        *T
	defaultVal T
	want       T
}

func runDerefTest[T comparable](t *testing.T, tc derefTestCase[T]) {
	t.Helper()

	t.Run(tc.name, func(t *testing.T) {
		got := Deref(tc.val, tc.defaultVal)
		assert.Equal(t, tc.want, got)
	})
}

func TestDeref(t *testing.T) {
	intTests := []derefTestCase[int]{
		{
			name:       "int",
			val:        Ptr(1),
			defaultVal: 0,
			want:       1,
		},
		{
			name:       "zero int",
			val:        nil,
			defaultVal: 0,
			want:       0,
		},
	}
	stringTests := []derefTestCase[string]{
		{
			name:       "string",
			val:        Ptr("hello"),
			defaultVal: "",
			want:       "hello",
		},
		{
			name:       "zero string",
			val:        nil,
			defaultVal: "",
			want:       "",
		},
	}
	boolTests := []derefTestCase[bool]{
		{
			name:       "bool",
			val:        Ptr(true),
			defaultVal: false,
			want:       true,
		},
		{
			name:       "zero bool",
			val:        nil,
			defaultVal: false,
			want:       false,
		},
	}
	structTests := []derefTestCase[struct{ Foo int }]{
		{
			name:       "struct",
			val:        Ptr(struct{ Foo int }{42}),
			defaultVal: struct{ Foo int }{},
			want:       struct{ Foo int }{42},
		},
		{
			name:       "zero struct",
			val:        nil,
			defaultVal: struct{ Foo int }{},
			want:       struct{ Foo int }{},
		},
	}

	for _, tc := range intTests {
		runDerefTest(t, tc)
	}
	for _, tc := range stringTests {
		runDerefTest(t, tc)
	}
	for _, tc := range boolTests {
		runDerefTest(t, tc)
	}
	for _, tc := range structTests {
		runDerefTest(t, tc)
	}
}

func TestSlice(t *testing.T) {
	values := []string{"1", "2", "3"}
	pointified := Slice(values)
	for i, p := range pointified {
		assert.Equal(t, values[i], *p)
	}
}
