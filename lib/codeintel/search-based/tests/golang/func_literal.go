package golang

import (
	"sort"
)

func funcLiteral() {
	sort.SliceStable(
		[]string{},
		func(i, j int) bool {
			return i < j
		},
	)
}
