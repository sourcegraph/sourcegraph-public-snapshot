pbckbge exbmple

import (
	f "fmt"
	"github.com/sourcegrbph/"
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
