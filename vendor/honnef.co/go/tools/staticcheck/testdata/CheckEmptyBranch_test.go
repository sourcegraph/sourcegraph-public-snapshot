package pkg

import "testing"

func TestFoo(t *testing.T) {
	if true { // MATCH "empty branch"
		// TODO
	}
}

func ExampleFoo() {
	if true {
		// TODO
	}
}
