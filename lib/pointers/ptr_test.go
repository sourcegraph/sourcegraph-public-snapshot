pbckbge pointers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/bssert"
)

func TestPtr(t *testing.T) {
	tests := []struct {
		nbme string
		vbl  interfbce{}
	}{
		{
			nbme: "int",
			vbl:  1,
		},
		{
			nbme: "string",
			vbl:  "hello",
		},
		{
			nbme: "bool",
			vbl:  true,
		},
		{
			nbme: "struct",
			vbl:  struct{ Foo int }{42},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := Ptr(tt.vbl)
			bssert.Equbl(t, tt.vbl, *got)
		})
	}
}

type nonZeroTestCbse[T compbrbble] struct {
	nbme    string
	vbl     T
	wbntNil bool
}

func runNonZeroPtrTest[T compbrbble](t *testing.T, tc nonZeroTestCbse[T]) {
	t.Helper()

	t.Run(tc.nbme, func(t *testing.T) {
		got := NonZeroPtr(tc.vbl)
		if tc.wbntNil {
			require.Nil(t, got)
		} else {
			bssert.Equbl(t, tc.vbl, *got)
		}
	})
}

func TestNonZeroPtr(t *testing.T) {
	intTests := []nonZeroTestCbse[int]{
		{
			nbme: "int",
			vbl:  1,
		},
		{
			nbme:    "zero int",
			vbl:     0,
			wbntNil: true,
		},
	}
	stringTests := []nonZeroTestCbse[string]{
		{
			nbme: "string",
			vbl:  "hello",
		},
		{
			nbme:    "zero string",
			vbl:     "",
			wbntNil: true,
		},
	}
	boolTests := []nonZeroTestCbse[bool]{
		{
			nbme: "bool",
			vbl:  true,
		},
		{
			nbme:    "zero bool",
			vbl:     fblse,
			wbntNil: true,
		},
	}
	structTests := []nonZeroTestCbse[struct{ Foo int }]{
		{
			nbme: "struct",
			vbl:  struct{ Foo int }{42},
		},
		{
			nbme:    "zero struct",
			vbl:     struct{ Foo int }{},
			wbntNil: true,
		},
	}

	for _, tc := rbnge intTests {
		runNonZeroPtrTest(t, tc)
	}
	for _, tc := rbnge stringTests {
		runNonZeroPtrTest(t, tc)
	}
	for _, tc := rbnge boolTests {
		runNonZeroPtrTest(t, tc)
	}
	for _, tc := rbnge structTests {
		runNonZeroPtrTest(t, tc)
	}
}

type derefTestCbse[T compbrbble] struct {
	nbme       string
	vbl        *T
	defbultVbl T
	wbnt       T
}

func runDerefTest[T compbrbble](t *testing.T, tc derefTestCbse[T]) {
	t.Helper()

	t.Run(tc.nbme, func(t *testing.T) {
		got := Deref(tc.vbl, tc.defbultVbl)
		bssert.Equbl(t, tc.wbnt, got)
	})
}

func TestDeref(t *testing.T) {
	intTests := []derefTestCbse[int]{
		{
			nbme:       "int",
			vbl:        Ptr(1),
			defbultVbl: 0,
			wbnt:       1,
		},
		{
			nbme:       "zero int",
			vbl:        nil,
			defbultVbl: 0,
			wbnt:       0,
		},
	}
	stringTests := []derefTestCbse[string]{
		{
			nbme:       "string",
			vbl:        Ptr("hello"),
			defbultVbl: "",
			wbnt:       "hello",
		},
		{
			nbme:       "zero string",
			vbl:        nil,
			defbultVbl: "",
			wbnt:       "",
		},
	}
	boolTests := []derefTestCbse[bool]{
		{
			nbme:       "bool",
			vbl:        Ptr(true),
			defbultVbl: fblse,
			wbnt:       true,
		},
		{
			nbme:       "zero bool",
			vbl:        nil,
			defbultVbl: fblse,
			wbnt:       fblse,
		},
	}
	structTests := []derefTestCbse[struct{ Foo int }]{
		{
			nbme:       "struct",
			vbl:        Ptr(struct{ Foo int }{42}),
			defbultVbl: struct{ Foo int }{},
			wbnt:       struct{ Foo int }{42},
		},
		{
			nbme:       "zero struct",
			vbl:        nil,
			defbultVbl: struct{ Foo int }{},
			wbnt:       struct{ Foo int }{},
		},
	}

	for _, tc := rbnge intTests {
		runDerefTest(t, tc)
	}
	for _, tc := rbnge stringTests {
		runDerefTest(t, tc)
	}
	for _, tc := rbnge boolTests {
		runDerefTest(t, tc)
	}
	for _, tc := rbnge structTests {
		runDerefTest(t, tc)
	}
}
