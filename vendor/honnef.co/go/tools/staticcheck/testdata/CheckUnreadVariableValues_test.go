package pkg

import "testing"

func fn() int { println(); return 0 }

func TestFoo(t *testing.T) {
	x := fn() // MATCH "never used"
	x = fn()
	println(x)
}

func ExampleFoo() {
	x := fn()
	x = fn()
	println(x)
}
