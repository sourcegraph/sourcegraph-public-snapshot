package pkg

func fn() int { println(); return 0 }

var x = func(arg int) { // MATCH "overwritten"
	arg = 1
	println(arg)
}

var y = func() {
	v := fn() // MATCH "never used"
	v = fn()
	println(v)
}

var z = func() {
	for {
		if true {
			println()
		}
		break // MATCH "the surrounding loop is unconditionally terminated"
	}
}
