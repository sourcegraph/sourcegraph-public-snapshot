package pkg

func fn() {
	var ch chan int
	<-ch
	_ = <-ch // MATCH /_ = <-ch/
	select {
	case <-ch:
	case _ = <-ch: // MATCH /_ = <-ch/
	}
	x := <-ch
	y, _ := <-ch, <-ch // MATCH /_ = <-ch/
	_, z := <-ch, <-ch // MATCH /_ = <-ch/
	_, _, _ = x, y, z
}
