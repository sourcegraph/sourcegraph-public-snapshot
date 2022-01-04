package golang

import "fmt"

func forStatement() {
	for hello := 0; hello < 10; hello++ {
		fmt.Println(hello)
	}

	for i, j := 0, 1; i < 10; i, j = i+1, j+2 {
		fmt.Println(i, j)
	}

	for n := range make(chan int, 1) {
		fmt.Println(n)
	}

	for i, e := range []string{} {
		fmt.Println(i, e)
	}
}
