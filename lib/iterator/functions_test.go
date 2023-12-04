package iterator_test

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func ExampleCollect() {
	it := iterator.From([]string{"Hello", "world"})
	v, err := iterator.Collect(it)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	// Output: [Hello world]
}
