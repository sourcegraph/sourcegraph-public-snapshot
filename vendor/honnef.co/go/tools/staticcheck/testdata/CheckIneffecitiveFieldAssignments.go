package pkg

import "fmt"

type T struct {
	X int
}

func (t T) Fn1() {
	t.X = 1 // MATCH /ineffective assignment to field X/
}

func (t T) Fn2() {
	t.X = 1
	fmt.Println(t)
}

func (t T) Fn3() {
	t.X = 1
	t.Fn4()
}

func (t T) Fn4() {
	t.X = 1
	println(t.X)
}

func (t T) Fn5() {
	fn1(&t)
	t.X = 1
}

func (t *T) Fn6() {
	t.X = 1
}

func fn1(*T) {}
