package main

import "fmt"

func main() {}

//       v simple-foo-x def
//       |      v simple-foo-y def
func foo(x int, y int) {
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

	//    vv simple-foo-c4 def
	const c4 = 4 // 💥

	//          vv simple-c1 ref
	//          ||  vv simple-c2 ref
	//          ||  ||  vv simple-c3 ref
	//          ||  ||  ||  vv simple-foo-c4 ref
	fmt.Println(c1, c2, c3, c4)
}

//    vv simple-c1 def
//    ||  vv simple-c2 def
const c1, c2 = 4, 4

const (
	c3 int = 4 // < "c3" simple-c3 def
)
