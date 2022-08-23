package main

import "fmt"

type Bar struct {
	z int
}

type Foo struct {
	x *int
	Y string
	Bar
	Bar2 Bar
	Bar3 *Bar
}

func main() {
	// this is comment

	x := 1234
	char := '1'
	aString := "hello\n"
	bool := true
	multilineString := `hello
	world
this is my poem` + aString

	var null_was_a_mistake *int
	null_was_a_mistake = nil

	foo := &Foo{
		x: &x,
		Bar: Bar{
			z: 43,
		},
	}

	fmt.Println(x, char, aString, bool, null_was_a_mistake, foo, multilineString)
}
