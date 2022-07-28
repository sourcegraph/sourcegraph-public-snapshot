package repos

func DefineNotify() {
	notify = func(ch chan struct{}) {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
