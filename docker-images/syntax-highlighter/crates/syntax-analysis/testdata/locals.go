package main

import (
	"fmt"
)

// Not actually local
var local = 10

func main() {
	local = 20
	local := 5
	something := func(unrelated int) int {
		superNested := func(deeplyNested int) int {
			return local + unrelated + deeplyNested
		}

		overwriteName := func(local int) int {
			return local + unrelated
		}

		return superNested(1) + overwriteName(1)
	}

	println(local, something)
}

func ifFn(local int) int {
	if local := 9; local < 0 {
		fmt.Println(local, "is negative")
	} else if local < 10 {
		fmt.Println(local, "has 1 digit")
	} else {
		fmt.Println(local, "has multiple digits")
	}
	return local
}

func switchFn(local int) int {
	switch x := 0; x {
	case 0:
		x := "local x"
		fmt.Println(x)
	case 1:
		fmt.Println(x)
	case x:
		fmt.Println("something")
	}
	return local
}

func forFn(local int) int {
	for i := 0; i < 3; i++ {
		fmt.Println(i)
	}
	return local

}

func constFunc() int {
	const LOCAL_CONST = 10
	return LOCAL_CONST
}

func assignmentFn(arg int) int {
	local := 0
	local = 1
	local[arg] = 2
	*local = 3
}

type MyStruct struct {
	field1 int
	field2 string
}

type MyInterface interface {
	method(param int) int
}

func (m *MyStruct) method(local int) int {
	return m.field1 + local
}

const MY_CONST int = 10
