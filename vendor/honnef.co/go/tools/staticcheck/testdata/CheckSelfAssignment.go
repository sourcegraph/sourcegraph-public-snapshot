package pkg

func fn(x int) {
	var z int
	var y int
	x = x // MATCH "self-assignment"
	y = y // MATCH "self-assignment"
	y, x, z = y, x, 1
	y = x
	_ = y
	_ = x
	_ = z
	func() {
		x := x
		println(x)
	}()
}

// MATCH:8 "self-assignment of y to y"
// MATCH:8 "self-assignment of x to x"
