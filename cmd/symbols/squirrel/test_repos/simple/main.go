package main

import "fmt"

func main() {}

//       v simple-foo-x def
//       |  v simple-foo-y def
func foo(x, y int) {
	//          v simple-foo-x ref
	//          |  v simple-foo-y ref
	fmt.Println(x, y)

	//        v simple-foo-anon-x def
	//        |      v simple-foo-anon-y def
	f := func(x int, y int) {
		//          v simple-foo-anon-x ref
		//          |  v simple-foo-anon-y ref
		fmt.Println(x, y)
	}

	g := f  // < "g" simple-foo-g def
	g(1, 2) // < "g" simple-foo-g ref

	//    vv simple-c1 def
	const c1 = 4 // 💥

	//          vv simple-c1 ref
	fmt.Println(c1)
}
