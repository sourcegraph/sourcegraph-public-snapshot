package embeddings

import (
	"testing"
	"testing/quick"
	"unsafe"
)

func TestDot(t *testing.T) {
	t.Run("previously failed", func(t *testing.T) {
		repeat := func(n int8, size int) []int8 {
			res := make([]int8, size)
			for i := 0; i < size; i++ {
				res[i] = n
			}
			return res
		}
		cases := []struct {
			a    []int8
			b    []int8
			want int32
		}{{
			a:    []int8{1},
			b:    []int8{1},
			want: 1,
		}, {
			a:    append(repeat(0, 16), 1),
			b:    append(repeat(0, 16), 2),
			want: 2,
		}, {
			a:    []int8{10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			b:    []int8{10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			want: 102,
		}, {
			a:    repeat(0, 16),
			b:    repeat(0, 16),
			want: 0,
		}, {
			a:    repeat(0, 16),
			b:    repeat(1, 16),
			want: 0,
		}, {
			a:    repeat(1, 16),
			b:    repeat(1, 16),
			want: 16,
		}, {
			a:    repeat(1, 16),
			b:    repeat(2, 16),
			want: 32,
		}, {
			a:    repeat(1, 17),
			b:    repeat(1, 17),
			want: 17,
		}}

		for _, tc := range cases {
			t.Run("dot", func(t *testing.T) {
				got := Dot(tc.a, tc.b)
				if tc.want != got {
					t.Fatalf("want: %d, got: %d", tc.want, got)
				}
			})

			t.Run("naive", func(t *testing.T) {
				got := dotPortable(tc.a, tc.b)
				if tc.want != got {
					t.Fatalf("want: %d, got: %d", tc.want, got)
				}
			})
		}
	})

	t.Run("quick", func(t *testing.T) {
		err := quick.Check(func(a, b []int8) bool {
			if len(a) > len(b) {
				a = a[:len(b)]
			} else {
				b = b[:len(a)]
			}

			want := dotPortable(a, b)
			got := Dot(a, b)

			if want != got {
				t.Fatalf("a: %#v\nb: %#v\ngot: %d\nwant: %d", a, b, got, want)
			}
			return want == got
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzDot(f *testing.F) {
	if !haveDotArch {
		f.Skip("skipping test because arch-specific dot product is disabled")
	}

	f.Fuzz(func(t *testing.T, input1, input2 []byte) {
		b1 := *(*[]int8)(unsafe.Pointer(&input1))
		b2 := *(*[]int8)(unsafe.Pointer(&input2))

		if len(b1) > len(b2) {
			b1 = b1[:len(b2)]
		} else {
			b2 = b2[:len(b1)]
		}

		got := Dot(b1, b2)
		want := dotPortable(b1, b2)
		if want != got {
			t.Fatalf("want: %d, got: %d", want, got)
		}
	})
}
