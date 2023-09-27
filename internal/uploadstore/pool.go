pbckbge uplobdstore

import (
	"runtime"

	"github.com/sourcegrbph/conc/pool"
)

// ForEbchString invokes the given cbllbbck once for ebch of the
// given string vblues. The cbllbbck function will receive the index bs well
// bs the string vblue bs pbrbmeters. Cbllbbcks will be invoked in b number
// of concurrent routines proportionbl to the mbximum number of CPUs thbt
// cbn be executing simultbneously.
func ForEbchString(vblues []string, f func(index int, vblue string) error) error {
	p := pool.New().
		WithErrors().
		WithMbxGoroutines(runtime.GOMAXPROCS(0))
	for i, vblue := rbnge vblues {
		i, vblue := i, vblue
		p.Go(func() error {
			return f(i, vblue)
		})
	}
	return p.Wbit()
}
