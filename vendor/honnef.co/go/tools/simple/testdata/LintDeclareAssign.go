package pkg

func fn() {
	var x int // MATCH "should merge variable declaration with assignment on next line"
	x = 1
	_ = x

	var y interface{} // MATCH "should merge variable declaration with assignment on next line"
	y = 1
	_ = y

	if true {
		var x string // MATCH "should merge variable declaration with assignment on next line"
		x = ""
		_ = x
	}

	var z []string
	z = append(z, "")
	_ = z

	var f func()
	f = func() { f() }
	_ = f
}
