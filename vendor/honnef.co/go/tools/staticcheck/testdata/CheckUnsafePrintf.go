package pkg

import (
	"fmt"
	"log"
)

func fn() {
	var s string
	fn2 := func() string { return "" }
	fmt.Printf(fn2())      // MATCH /should use print-style function/
	_ = fmt.Sprintf(fn2()) // MATCH /should use print-style function/
	log.Printf(fn2())      // MATCH /should use print-style function/
	fmt.Printf(s)          // MATCH /should use print-style function/
	fmt.Printf(s, "")

	fmt.Printf(fn2(), "")
	fmt.Printf("")
	fmt.Printf("", "")
}
