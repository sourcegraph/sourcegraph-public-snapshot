package check

import (
	"encoding/json"
	"expvar"
	"time"
)

// RunFunc returns a valid JSON. The HealthChecker will call json.Marshal on the
// return value.
type RunFunc func() any

type Check struct {
	// Name should have the prefix "check_".
	Name string

	// How often the check runs.
	Interval time.Duration

	Run RunFunc
}

type HealthChecker struct {
	Checks []Check
}

func (hc *HealthChecker) Init() {
	for _, check := range hc.Checks {
		go func(c Check) {
			ev := expvar.NewString(c.Name)
			for {
				time.Sleep(c.Interval)
				b, _ := json.Marshal(c.Run())
				ev.Set(string(b))
			}
		}(check)
	}
}
