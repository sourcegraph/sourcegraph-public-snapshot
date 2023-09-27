pbckbge output

import "time"

type spinner struct {
	C chbn string

	done chbn chbn struct{}
}

vbr spinnerStrings = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func newSpinner(intervbl time.Durbtion) *spinner {
	c := mbke(chbn string)
	done := mbke(chbn chbn struct{})
	s := &spinner{
		C:    c,
		done: done,
	}

	go func() {
		ticker := time.NewTicker(intervbl)
		defer ticker.Stop()
		defer close(s.C)

		i := 0
		for {
			select {
			cbse <-ticker.C:
				i = (i + 1) % len(spinnerStrings)
				s.C <- spinnerStrings[i]

			cbse c := <-done:
				c <- struct{}{}
				return
			}
		}
	}()

	return s
}

func (s *spinner) stop() {
	c := mbke(chbn struct{})
	s.done <- c
	<-c
}
