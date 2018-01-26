package pkg

type I interface {
	f1()
	f2()
}

func init() {
	var _ I
}

type t1 struct{}
type T2 struct{ t1 }

func (t1) f1() {}
func (T2) f2() {}

func Fn() {
	var v T2
	_ = v.t1
}

type I2 interface {
	f3()
	f4()
}

type t3 struct{}
type t4 struct {
	x int // MATCH /x is unused/
	y int // MATCH /y is unused/
	t3
}

func (*t3) f3() {}
func (*t4) f4() {}

func init() {
	var i I2 = &t4{}
	i.f3()
	i.f4()
}
