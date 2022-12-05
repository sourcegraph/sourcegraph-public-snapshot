package iterator_test

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func ExampleIterator() {
	x := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	it := iterator.New(func() ([]int, error) {
		if len(x) == 0 {
			return nil, nil
		}
		y := x[:2]
		x = x[2:]
		return y, nil
	})

	for it.Next() {
		fmt.Printf("%d ", it.Current())
	}

	if it.Err() != nil {
		fmt.Println(it.Err())
	}

	// Output: 1 2 3 4 5 6 7 8 9 10
}
