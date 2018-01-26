package pkg

func fn() {
	for false {
	}
	for true { // MATCH /should use for/
	}
	for {
	}
	for i := 0; true; i++ {
	}
}
