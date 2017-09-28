package pkg

func fn1() bool { return false }
func fn2() bool { return false }

func fn() {
	type T bool
	var x T
	const t T = false
	if x == t {
	}
	if fn1() == true { // MATCH "simplified to fn1()"
	}
	if fn1() != true { // MATCH "simplified to !fn1()"
	}
	if fn1() == false { // MATCH "simplified to !fn1()"
	}
	if fn1() != false { // MATCH "simplified to fn1()"
	}
	if fn1() && (fn1() || fn1()) || (fn1() && fn1()) == true { // MATCH "simplified to (fn1() && fn1())"
	}

	if (fn1() && fn2()) == false { // MATCH "simplified to !(fn1() && fn2())"
	}

	var y bool
	for y != true { // MATCH /simplified to !y/
	}
	if !y == true { // MATCH /simplified to !y/
	}
	if !y == false { // MATCH /simplified to y/
	}
	if !y != true { // MATCH /simplified to y/
	}
	if !y != false { // MATCH /simplified to !y/
	}
	if !!y == false { // MATCH /simplified to !y/
	}
	if !!!y == false { // MATCH /simplified to y/
	}
	if !!y == true { // MATCH /simplified to y/
	}
	if !!!y == true { // MATCH /simplified to !y/
	}
	if !!y != true { // MATCH /simplified to !y/
	}
	if !!!y != true { // MATCH /simplified to y/
	}
	if !y == !false { // not matched because we expect true/false on one side, not !false
	}

	var z interface{}
	if z == true {
	}
}
