package pkg

func fn() {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; i++ { // MATCH /variable in loop condition never changes/
		}
	}
}

// M_ATCH:5 /j < 10 is always true for all possible values/
