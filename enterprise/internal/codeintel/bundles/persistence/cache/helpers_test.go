package cache

import "time"

var TestTickDuration = time.Millisecond * 25

// waitForCondition will block a short time while the invocation of f returns
// false. This allows cache-internal timers to have a chance to be scheduled
// before moving on to the test assertions that depend on this condition.
func waitForCondition(f func() bool) {
	for attempts := 5; attempts >= 0; attempts-- {
		if f() {
			return
		}

		<-time.After(TestTickDuration)
	}
}
