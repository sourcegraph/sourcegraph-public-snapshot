package pkg

func fn() {
	var ch chan int
	for range ch {
		defer println() // MATCH /defers in this range loop/
	}
}
