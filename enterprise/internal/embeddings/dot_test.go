package embeddings

import (
	"testing"
	"testing/quick"
)

func TestDot(t *testing.T) {
	t.Run("edge cases", func(t *testing.T) {
		repeat := func(n int8, size int) []int8 {
			res := make([]int8, size)
			for i := 0; i < size; i++ {
				res[i] = n
			}
			return res
		}

		interval := func(start int8, size int) []int8 {
			res := make([]int8, size)
			for i := 0; i < size; i++ {
				res[i] = start + int8(i)
			}
			return res
		}

		cases := []struct {
			a []int8
			b []int8
		}{
			{[]int8{}, []int8{}},
			{[]int8{1}, []int8{1}},
			{append(repeat(0, 16), 1), append(repeat(0, 16), 2)},
			{repeat(0, 16), repeat(0, 16)},
			{repeat(0, 16), repeat(1, 16)},
			{repeat(1, 16), repeat(1, 16)},
			{repeat(1, 16), repeat(2, 16)},
			{repeat(1, 17), repeat(1, 17)},

			// Some non-constant tests
			{interval(0, 99), repeat(0, 99)},
			{interval(0, 99), repeat(1, 99)},

			// A couple of large ones to ensure no weird behavior at scale
			{repeat(1, 1000000), repeat(1, 1000000)},
			{repeat(1, 1000000), repeat(2, 1000000)},

			// This will come very close to overflowing an int32.
			// Make sure nothing crashes.
			{repeat(127, 133000), repeat(127, 133000)},

			// This will overflow an int32 and return garbage.
			// Just make sure nothing crashes.
			{repeat(127, 134000), repeat(127, 134000)},

			// These will overflow if we don't multiply into larger ints
			{repeat(127, 40), repeat(127, 40)},
			{repeat(-128, 40), repeat(-128, 40)},
			{repeat(-128, 40), repeat(127, 40)},
		}

		for _, tc := range cases {
			t.Run("dot", func(t *testing.T) {
				got := Dot(tc.a, tc.b)
				want := dotNaive(tc.a, tc.b)
				if want != got {
					t.Fatalf("want: %d, got: %d", want, got)
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

			want := dotNaive(a, b)
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

func dotNaive(a, b []int8) int32 {
	sum := int32(0)
	for i := 0; i < len(a); i++ {
		sum += int32(a[i]) * int32(b[i])
	}
	return sum
}
