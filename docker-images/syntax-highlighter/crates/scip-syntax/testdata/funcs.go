package example

import (
	f "fmt"
)

func Something() {
	y := ", world"
	f.Println("hello", y)
}

func Another() {
	Something()
	if true {
		x := true
	}
	if true {
		x := true
		if true {
			x := true
		}
	}
	if true {
		x := true
	}
}
