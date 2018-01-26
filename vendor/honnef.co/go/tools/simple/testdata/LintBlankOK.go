package pkg

func fn() {
	var m map[int]int
	var ch chan int
	var fn func() (int, bool)

	x, _ := m[0] // MATCH "should write x := m[0] instead of x, _ := m[0]"
	x, _ = <-ch  // MATCH "should write x = <-ch instead of x, _ = <-ch"
	x, _ = fn()
	_ = x
}
