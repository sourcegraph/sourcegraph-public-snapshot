package pkg

func a() { // MATCH /a is unused/
	b()
}

func b() { // MATCH /b is unused/
	a()
}
