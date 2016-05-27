package backoff

import (
	"math"
	"math/rand"
	"time"
)

//Backoff is a time.Duration counter. It starts at Min.
//After every call to Duration() it is  multiplied by Factor.
//It is capped at Max. It returns to Min on every call to Reset().
//Used in conjunction with the time package.
//
// Backoff is not threadsafe, but the ForAttempt method can be
// used concurrently if non-zero values for Factor, Max, and Min
// are set on the Backoff shared among threads.
type Backoff struct {
	//Factor is the multiplying factor for each increment step
	attempt, Factor float64
	//Jitter eases contention by randomizing backoff steps
	Jitter bool
	//Min and Max are the minimum and maximum values of the counter
	Min, Max time.Duration
}

//Returns the current value of the counter and then
//multiplies it Factor
func (b *Backoff) Duration() time.Duration {
	d := b.ForAttempt(b.attempt)
	b.attempt++
	return d
}

// ForAttempt returns the duration for a specific attempt. This is useful if
// you have a large number of independent Backoffs, but don't want use
// unnecessary memory storing the Backoff parameters per Backoff. The first
// attempt should be 0.
//
// ForAttempt is threadsafe iff non-zero values for Factor, Max, and Min
// are set before any calls to ForAttempt are made.
func (b *Backoff) ForAttempt(attempt float64) time.Duration {
	//Zero-values are nonsensical, so we use
	//them to apply defaults
	if b.Min == 0 {
		b.Min = 100 * time.Millisecond
	}
	if b.Max == 0 {
		b.Max = 10 * time.Second
	}
	if b.Factor == 0 {
		b.Factor = 2
	}
	//calculate this duration
	dur := float64(b.Min) * math.Pow(b.Factor, attempt)
	if b.Jitter == true {
		dur = rand.Float64()*(dur-float64(b.Min)) + float64(b.Min)
	}
	//cap!
	if dur > float64(b.Max) {
		return b.Max
	}
	//return as a time.Duration
	return time.Duration(dur)
}

//Resets the current value of the counter back to Min
func (b *Backoff) Reset() {
	b.attempt = 0
}

//Get the current backoff attempt
func (b *Backoff) Attempt() float64 {
	return b.attempt
}
