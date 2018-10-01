package goroutine

import "testing"

func TestGo(t *testing.T) {
	done := make(chan bool)
	Go(func() {
		defer func() {
			done <- true
		}()
		panic("will be caught and test will not fail")
	})
	<-done
}
