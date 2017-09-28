package pkg

type T struct{}

func fn1(_ *T) {}

func fn2() {
	t1 := &T{}
	fn1(&*t1) // MATCH /will not copy/
	fn1(*&t1) // MATCH /will not copy/

	_Cvar_something := &T{}
	fn1(&*_Cvar_something)
}
