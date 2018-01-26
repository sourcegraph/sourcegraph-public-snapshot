package pkg

import _ "fmt"

type t1 struct{} // MATCH /t1 is unused/
type t2 struct{}
type t3 struct{}

var _ = t2{}

func fn1() { // MATCH /fn1 is unused/
	_ = t1{}
	var _ = t1{}
}

func fn2() {
	_ = t3{}
}

func init() {
	fn2()
}

func _() {}

type _ struct{}
