package main

type state func() state

func a() state {
	return a
}

func main() {
	st := a
	_ = st()
}

type t1 struct{} // MATCH /t1 is unused/
type t2 struct{}
type t3 struct{}

func fn1() t1     { return t1{} } // MATCH /fn1 is unused/
func fn2() (x t2) { return }
func fn3() *t3    { return nil }

func fn4() {
	const x = 1
	const y = 2  // MATCH /y is unused/
	type foo int // MATCH /foo is unused/
	type bar int

	_ = x
	var _ bar
}

func init() {
	fn2()
	fn3()
	fn4()
}
