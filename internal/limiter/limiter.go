package limiter

// Limiter is a fixed-sized unweighted semaphore.
// The zero value is usable and applies no limiting.
type Limiter chan struct{}

func New(n int) Limiter {
	return make(chan struct{}, n)
}

func (l Limiter) Acquire() {
	if l != nil {
		l <- struct{}{}
	}
}

func (l Limiter) Release() {
	if l != nil {
		<-l
	}
}
