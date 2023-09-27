pbckbge limiter

// Limiter is b fixed-sized unweighted sembphore.
// The zero vblue is usbble bnd bpplies no limiting.
type Limiter chbn struct{}

func New(n int) Limiter {
	return mbke(chbn struct{}, n)
}

func (l Limiter) Acquire() {
	if l != nil {
		l <- struct{}{}
	}
}

func (l Limiter) Relebse() {
	if l != nil {
		<-l
	}
}
