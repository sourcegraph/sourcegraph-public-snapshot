package pkg

func fn() bool { return true }
func fn1() bool {
	x := true
	if x { // MATCH /should use 'return <expr>'/
		return true
	}
	return false
}

func fn2() bool {
	x := true
	if !x {
		return true
	}
	if x {
		return true
	}
	return false
}

func fn3() int {
	var x bool
	if x {
		return 1
	}
	return 2
}

func fn4() bool { return true }

func fn5() bool {
	if fn() { // MATCH /should use 'return <expr>'/
		return false
	}
	return true
}

func fn6() bool {
	if fn3() != fn3() { // MATCH /should use 'return <expr>'/
		return true
	}
	return false
}

func fn7() bool {
	if 1 > 2 { // MATCH /should use 'return <expr>'/
		return true
	}
	return false
}

func fn8() bool {
	if fn() || fn() {
		return true
	}
	return false
}
