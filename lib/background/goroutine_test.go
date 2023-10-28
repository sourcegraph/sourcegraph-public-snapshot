package background

import "testing"

type closeOnError chan bool

func (e closeOnError) Error() string {
	close(e)
	return "will be caught and test will not fail"
}

func TestGo(t *testing.T) {
	done := make(chan bool)
	Go(func() {
		// The recover handler for Go logs the value we pass to panic. When it
		// does log this closeOnError.Error is called which will close
		// done. This is to ensure we actually run the recover path before
		// returning from the test.
		panic(closeOnError(done))
	})
	<-done
}
