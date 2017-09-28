package pkg

func fn() {
	if true { // MATCH "empty branch"
	}
	if true { // MATCH "empty branch"
	} else { // MATCH "empty branch"
	}
	if true {
		println()
	}

	if true {
		println()
	} else { // MATCH "empty branch"
	}

	if true { // MATCH "empty branch"
		// TODO handle error
	}

	if true {
	} else {
		println()
	}

	if true {
	} else if false { // MATCH "empty branch"
	}
}
