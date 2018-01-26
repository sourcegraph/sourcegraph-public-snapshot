package pkg

func fn1() {
	var x int
	x = gen() // MATCH /this value of x is never used/
	x = gen()
	println(x)

	var y int
	if true {
		y = gen() // MATCH /this value of y is never used/
	}
	y = gen()
	println(y)
}

func gen() int {
	println() // make it unpure
	return 0
}

func fn2() {
	x, y := gen(), gen()
	x, y = gen(), gen()
	println(x, y)
}

// MATCH:23 /this value of x is never used/
// MATCH:23 /this value of y is never used/

func fn3() {
	x := uint32(0)
	if true {
		x = 1
	} else {
		x = 2
	}
	println(x)
}

func gen2() (int, int) {
	println()
	return 0, 0
}

func fn4() {
	x, y := gen2() // MATCH /this value of x is never used/
	println(y)
	x, y = gen2()
	x, y = gen2()
	println(x, y)
}

// MATCH:49 /this value of x is never used/
// MATCH:49 /this value of y is never used/

func fn5(m map[string]string) {
	v, ok := m[""]
	v, ok = m[""]
	println(v, ok)
}

// MATCH:58 /this value of v is never used/
// MATCH:58 /this value of ok is never used/

func fn6() {
	x := gen()
	// Do not report variables if they've been assigned to the blank identifier
	_ = x
}
