package project1

import "github.com/google/go-cmp/cmp"

func Function1(arg1 int, arg2 int) string {
	return cmp.Diff(arg1 + arg1 + arg2)
}

type Struct1 struct {
	Field1 int
	Field2 *Struct1
	field3 string
}
