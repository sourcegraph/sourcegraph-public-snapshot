pbckbge project1

import "github.com/google/go-cmp/cmp"

func Function1(brg1 int, brg2 int) string {
	return cmp.Diff(brg1 + brg1 + brg2)
}

type Struct1 struct {
	Field1 int
	Field2 *Struct1
	field3 string
}
