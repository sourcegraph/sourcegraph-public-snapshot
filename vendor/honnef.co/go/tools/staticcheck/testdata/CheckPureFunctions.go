package pkg

import (
	"context"
	"net/http"
	"strings"
)

func fn1() {
	strings.Replace("", "", "", 1) // MATCH /is a pure function but its return value is ignored/
	foo(1, 2)                      // MATCH /is a pure function but its return value is ignored/
	bar(1, 2)
}

func fn2() {
	r, _ := http.NewRequest("GET", "/", nil)
	r.WithContext(context.Background()) // MATCH /is a pure function but its return value is ignored/
}

func foo(a, b int) int { return a + b }
func bar(a, b int) int {
	println(a + b)
	return a + b
}

func empty() {}

func fn3() {
	empty()
}
