package main

import "fmt"

func main() {}

func foo(x int, y int) {
	//   ^ simple-foo-x def
	//          ^ simple-foo-y def
	fmt.Println(x, y)
	//          ^ simple-foo-x ref
	//             ^ simple-foo-y ref

	f := func(x int, y int) {
		//    ^ simple-foo-anon-x def
		//           ^ simple-foo-anon-y def
		fmt.Println(x, y)
		//          ^ simple-foo-anon-x ref
		//             ^ simple-foo-anon-y ref
	}

	g := f  // < "g" simple-foo-g def
	g(1, 2) // < "g" simple-foo-g ref
}
