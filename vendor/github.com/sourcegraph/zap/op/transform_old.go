package op

import (
	"fmt"

	"github.com/go-kit/kit/log"
)

// oldTransform transforms the concurrent ops a and b into a1 and b1
// such that compose(a, b1) == compose(b1, a).
func oldTransform(logger log.Logger, a, b Ops) (a1, b1 Ops, err error) {
	// Merge inputs to canonicalize/simplify them. This also creates a
	// copy of them so we don't modify the inputs.
	a, err = Merge(logger, a)
	if err != nil {
		return nil, nil, err
	}
	b, err = Merge(logger, b)
	if err != nil {
		return nil, nil, err
	}

	transformAB := func(a Op, b Ops) (a1, b1 Ops, err error) {
		if a == nil {
			return nil, b, nil
		}
		return a.Transform(b)
	}

	// replaceInSlice removes x[i] and inserts v's elements in its
	// place.
	replaceInSlice := func(x []Op, i *int, v []Op) []Op {
		switch len(v) {
		case 0:
			x[*i] = nil
		case 1:
			x[*i] = v[0]
		default:
			x = append(x[:*i], append(v, x[*i+1:]...)...)
			*i = *i + len(v) - 1
		}
		return x
	}

	for ai := 0; ai < len(a); ai++ {
		fmt.Printf("transform(a=%v   b=%v) ==> ", a[ai], b)
		a1, b1, err := transformAB(a[ai], b)
		if err != nil {
			return nil, nil, err
		}
		fmt.Printf("a1=%v   b1=%v\n", a1, b1)
		a = replaceInSlice(a, &ai, a1)
		b = b1
	}

	fmt.Println("---")
	fmt.Printf("a=%v\n", a)
	fmt.Printf("b=%v\n", b)

	a, err = Merge(logger, a)
	if err != nil {
		return nil, nil, err
	}
	b, err = Merge(logger, b)
	if err != nil {
		return nil, nil, err
	}

	return a, b, nil
}
