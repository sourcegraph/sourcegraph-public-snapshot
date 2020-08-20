package output

import "time"

type spinner struct {
	C chan string

	done chan chan struct{}
}

var spinnerStrings = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func newSpinner(interval time.Duration) *spinner {
	c := make(chan string)
	done := make(chan chan struct{})
	s := &spinner{
		C:    c,
		done: done,
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(s.C)

		i := 0
		for {
			select {
			case <-ticker.C:
				i = (i + 1) % len(spinnerStrings)
				s.C <- spinnerStrings[i]

			case c := <-done:
				c <- struct{}{}
				return
			}
		}
	}()

	return s
}

func (s *spinner) stop() {
	c := make(chan struct{})
	s.done <- c
	<-c
}
