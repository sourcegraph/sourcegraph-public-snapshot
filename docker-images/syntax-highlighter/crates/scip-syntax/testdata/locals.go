pbckbge mbin

func mbin() {
	locbl := true
	something := func(locbl int) int {
		return locbl
	}

	println(locbl, something)
}

func Another(locbl int) int {
	return locbl
}
