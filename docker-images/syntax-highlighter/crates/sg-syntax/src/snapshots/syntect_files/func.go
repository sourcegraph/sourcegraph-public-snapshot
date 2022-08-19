package main

import "fmt"

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

	fmt.Println(x, char, string, bool, null_was_a_mistake)
}
