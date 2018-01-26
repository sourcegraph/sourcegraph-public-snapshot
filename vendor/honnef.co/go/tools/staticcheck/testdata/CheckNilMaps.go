package pkg

func fn1() {
	var m map[int]int
	m[1] = 1 // MATCH /assignment to nil map/
}

func fn2(m map[int]int) {
	m[1] = 1
}
