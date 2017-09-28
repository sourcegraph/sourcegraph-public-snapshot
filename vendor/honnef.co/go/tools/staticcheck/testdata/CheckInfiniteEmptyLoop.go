package pkg

func fn2() bool { return true }

func fn() {
	for { // MATCH /this loop will spin/
	}

	for fn2() {
	}

	for {
		break
	}

	for true { // MATCH "loop condition never changes"
	}
}

// MATCH:16 "this loop will spin"
